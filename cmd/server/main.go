package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"seek/internal/appdb"
	"seek/internal/config"
	"seek/internal/eventcatalog"
	"seek/internal/eventstore"
	"seek/internal/features/period"
	"seek/internal/features/schedule"
	"seek/internal/features/student"
	"seek/internal/features/teacher"
	"seek/internal/httpserver"
	"seek/internal/natsbus"
	"seek/internal/viewstore"
)

type runOptions struct {
	migrateOnly bool
	seedOnly    bool
}

type appComponents struct {
	periodReadModel   *period.ReadModel
	scheduleReadModel *schedule.ReadModel
	studentReadModel  *student.ReadModel
	teacherReadModel  *teacher.ReadModel
	checkpointer      eventstore.Checkpointer
}

func main() {
	opts := parseOptions()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(ctx, stop, config.Load(), opts, logger); err != nil {
		logger.Error("run app", "err", err)
		os.Exit(1)
	}
}

func parseOptions() runOptions {
	migrateOnly := flag.Bool("migrate-only", false, "run database migrations and exit")
	seedOnly := flag.Bool("seed-only", false, "run seed tasks and exit")
	flag.Parse()
	return runOptions{migrateOnly: *migrateOnly, seedOnly: *seedOnly}
}

func run(ctx context.Context, stop context.CancelFunc, cfg config.Config, opts runOptions, logger *slog.Logger) error {
	db, err := appdb.Open(ctx, cfg.SQLitePath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer closeDB(db, logger)

	if opts.migrateOnly {
		logger.Info("sqlite migrations complete", "path", cfg.SQLitePath)
		return nil
	}
	if opts.seedOnly {
		logger.Info("seed requested; no seed tasks are currently defined")
		return nil
	}

	orisunStore, err := startEventStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer orisunStore.Close(context.Background())

	bus, err := natsbus.FromConn(orisunStore.NATSConnection())
	if err != nil {
		return fmt.Errorf("borrow embedded orisun nats: %w", err)
	}
	defer bus.Close()

	viewStore := newViewStore(bus, logger)
	components := newAppComponents(db, orisunStore, cfg, logger)
	handlers, err := startEventHandlers(ctx, eventHandlerFactories(orisunStore, bus, cfg, components, logger))
	if err != nil {
		return err
	}
	defer stopEventHandlers(handlers)

	app := httpserver.Server{
		// Accounts:       components.accountCommands,
		// Sessions:       components.sessionManager,
		// AuthUsers:   components.authUsers,
		Periods:        components.periodReadModel,
		Schedules:      components.scheduleReadModel,
		Students:       components.studentReadModel,
		Teachers:       components.teacherReadModel,
		EventSaver:     orisunStore,
		EventRetriever: orisunStore,
		// ProfileStorage: components.profileStorage,
		Subscriber:  bus,
		ViewStore:   viewStore,
		Development: cfg.DevelopmentCookie,
	}
	return serveHTTP(ctx, stop, cfg.Port, app.Routes(), logger)
}

func closeDB(db *appdb.DB, logger *slog.Logger) {
	if err := db.Close(); err != nil {
		logger.Error("close sqlite", "err", err)
	}
}

func startEventStore(ctx context.Context, cfg config.Config) (*eventstore.EmbeddedOrisun, error) {
	store, err := eventstore.StartEmbeddedOrisun(ctx, eventstore.EmbeddedConfig{
		Boundary:     cfg.OrisunBoundary,
		SQLiteDir:    cfg.OrisunSQLiteDir,
		NATSStoreDir: eventstore.PollingStoreDir(),
		LogLevel:     "info",
	})
	if err != nil {
		return nil, fmt.Errorf("start embedded orisun: %w", err)
	}
	if err := store.EnsureBoundaryIndexes(ctx, eventcatalog.BoundaryIndexes()); err != nil {
		store.Close(context.Background())
		return nil, fmt.Errorf("ensure orisun indexes: %w", err)
	}
	return store, nil
}

func newViewStore(bus *natsbus.Bus, logger *slog.Logger) viewstore.Store {
	store, err := viewstore.NewNATSStore(bus.Conn(), "go-starter-view-state", 5*time.Minute)
	if err != nil {
		logger.Warn("using in-memory view store fallback", "err", err)
		return viewstore.NewMemoryStore()
	}
	return store
}

func newAppComponents(db *appdb.DB, store *eventstore.EmbeddedOrisun, cfg config.Config, logger *slog.Logger) appComponents {
	periodReadModel := period.NewReadModel(db)
	scheduleReadModel := schedule.NewReadModel(db)
	studentReadModel := student.NewReadModel(db)
	teacherReadModel := teacher.NewReadModel(db)

	return appComponents{
		periodReadModel:   periodReadModel,
		scheduleReadModel: scheduleReadModel,
		studentReadModel:  studentReadModel,
		teacherReadModel:  teacherReadModel,
		checkpointer:      eventstore.NewSQLiteCheckpointer(db),
	}
}

func serveHTTP(ctx context.Context, stop context.CancelFunc, port string, handler http.Handler, logger *slog.Logger) error {
	server := &http.Server{Addr: ":" + port, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("starting server", "addr", "http://localhost:"+port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
