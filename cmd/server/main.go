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
	"seek/internal/auth"
	"seek/internal/config"
	"seek/internal/dbsql"
	"seek/internal/email"
	"seek/internal/eventcatalog"
	"seek/internal/eventstore"
	educatorEvents "seek/internal/features/educators/events"
	iepService "seek/internal/features/iep_services/events"
	period "seek/internal/features/periods/events"
	periodSchedule "seek/internal/features/periods_schedules/events"
	periodStudent "seek/internal/features/periods_students/events"
	profile "seek/internal/features/profiles/events"
	schedule "seek/internal/features/schedules/events"
	student "seek/internal/features/students/events"
	teacher "seek/internal/features/teachers/events"
	"seek/internal/httpserver"
	"seek/internal/natsbus"
	"seek/internal/protectedpii"
	"seek/internal/storage"
	"seek/internal/viewstore"

	"zombiezen.com/go/sqlite"
)

type runOptions struct {
	migrateOnly     bool
	seedOnly        bool
	resetReadModels bool
}

type appComponents struct {
	sessionManager          *auth.SessionManager
	authUsers               *auth.AuthUserStore
	verifications           *auth.VerificationStore
	educatorReadModel       *educatorEvents.ReadModel
	iepServiceReadModel     *iepService.ReadModel
	periodReadModel         *period.ReadModel
	profileReadModel        *profile.ReadModel
	scheduleReadModel       *schedule.ReadModel
	studentReadModel        *student.ReadModel
	teacherReadModel        *teacher.ReadModel
	periodScheduleReadModel *periodSchedule.ReadModel
	periodStudentReadModel  *periodStudent.ReadModel
	checkpointer            eventstore.Checkpointer
	emailSender             email.Sender
	profileStorage          profile.ObjectStore
	piiKeys                 *auth.SubjectPiiKeyStore
	accountDeletion         *auth.AccountDataDeletionStore
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
	resetReadModels := flag.Bool("reset-read-models", false, "reset SQLite read models and event-handler checkpoints")
	flag.Parse()
	return runOptions{migrateOnly: *migrateOnly, seedOnly: *seedOnly, resetReadModels: *resetReadModels}
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
	if opts.resetReadModels {
		if err := resetReadModels(ctx, db); err != nil {
			return fmt.Errorf("reset read models: %w", err)
		}
		logger.Info("read models and event-handler checkpoints reset", "path", cfg.SQLitePath)
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
		Sessions:            components.sessionManager,
		AuthUsers:           components.authUsers,
		PIIKeys:             components.piiKeys,
		PasswordCredentials: components.authUsers,
		Verifications:       components.verifications,
		Educators:           components.educatorReadModel,
		IEPServices:         components.iepServiceReadModel,
		Profiles:            components.profileReadModel,
		Periods:             components.periodReadModel,
		Schedules:           components.scheduleReadModel,
		Students:            components.studentReadModel,
		Teachers:            components.teacherReadModel,
		PeriodsSchedules:    components.periodScheduleReadModel,
		PeriodsStudents:     components.periodStudentReadModel,
		EventSaver:          orisunStore,
		EventRetriever:      orisunStore,
		ProfileStorage:      components.profileStorage,
		Subscriber:          bus,
		ViewStore:           viewStore,
		Development:         cfg.DevelopmentCookie,
		Logger:              logger,
	}
	return serveHTTP(ctx, stop, cfg.Port, app.Routes(), logger)
}

func closeDB(db *appdb.DB, logger *slog.Logger) {
	if err := db.Close(); err != nil {
		logger.Error("close sqlite", "err", err)
	}
}

func resetReadModels(ctx context.Context, db *appdb.DB) error {
	return db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		if err := dbsql.OnceResetReadModelAuthSessions(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelAuthAccounts(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelAuthVerifications(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelProfiles(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelIepservices(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelPeriods(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelSchedules(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelStudents(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelTeachers(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelPeriodsSchedules(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelPeriodsStudents(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelAuthUsers(conn); err != nil {
			return err
		}
		return dbsql.OnceResetEventHandlerCheckpoints(conn)
	})
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
	store, err := viewstore.NewNATSStore(bus.Conn(), "seek-view-state", 5*time.Minute)
	if err != nil {
		logger.Warn("using in-memory view store fallback", "err", err)
		return viewstore.NewMemoryStore()
	}
	return store
}

func newAppComponents(db *appdb.DB, store *eventstore.EmbeddedOrisun, cfg config.Config, logger *slog.Logger) appComponents {
	authUsers := auth.NewAuthUserStore(db)
	sessionManager := auth.NewSessionManager(db, authUsers, !cfg.DevelopmentCookie)
	verifications := auth.NewVerificationStore(db)
	educatorReadModel := educatorEvents.NewReadModel(db)
	iepServiceReadModel := iepService.NewReadModel(db)
	profileReadModel := profile.NewReadModel(db)
	periodReadModel := period.NewReadModel(db)
	scheduleReadModel := schedule.NewReadModel(db)
	studentReadModel := student.NewReadModel(db)
	teacherReadModel := teacher.NewReadModel(db)
	periodScheduleReadModel := periodSchedule.NewReadModel(db)
	periodStudentReadModel := periodStudent.NewReadModel(db)
	profileStorage := storage.NewLocalProvider(cfg.UploadDir, cfg.UploadBaseURL)
	piiKeys := auth.NewSubjectPiiKeyStore(db, protectedpii.FromEnv())
	accountDeletion := auth.NewAccountDataDeletionStore(db, profileStorage)

	return appComponents{
		sessionManager:          sessionManager,
		authUsers:               authUsers,
		verifications:           verifications,
		educatorReadModel:       educatorReadModel,
		iepServiceReadModel:     iepServiceReadModel,
		profileReadModel:        profileReadModel,
		periodReadModel:         periodReadModel,
		scheduleReadModel:       scheduleReadModel,
		studentReadModel:        studentReadModel,
		teacherReadModel:        teacherReadModel,
		periodScheduleReadModel: periodScheduleReadModel,
		periodStudentReadModel:  periodStudentReadModel,
		checkpointer:            eventstore.NewSQLiteCheckpointer(db),
		emailSender:             email.LogSender{Logger: logger},
		profileStorage:          profileStorage,
		piiKeys:                 piiKeys,
		accountDeletion:         accountDeletion,
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
