package appcore

import (
	"context"
	"fmt"
	"log/slog"

	"seek/internal/auth"
	"seek/internal/config"
	"seek/internal/email"
	"seek/internal/eventstore"
	caseloadStudentsEvents "seek/internal/features/caseload_students/events"
	educatorEvents "seek/internal/features/educators/events"
	educatorPeriodEvents "seek/internal/features/educators_periods/events"
	iepServiceEvents "seek/internal/features/iepservices/events"
	periodEvents "seek/internal/features/periods/events"
	profileEvents "seek/internal/features/profiles/events"
	studentEvents "seek/internal/features/students/events"
	periodStudentEvents "seek/internal/features/students_periods/events"
	"seek/internal/natsbus"
)

type eventHandler interface {
	StartSubscribing(context.Context) error
	StopSubscribing()
}

type eventHandlerFactory struct {
	name   string
	create func() (eventHandler, error)
}

func EventHandlerFactories(
	store *eventstore.EmbeddedOrisun,
	bus *natsbus.Bus,
	cfg config.Config,
	readModels *ReadModelContainer,
	authUsers *auth.AuthUserStore,
	checkpointer eventstore.Checkpointer,
	emailSender email.Sender,
	piiKeys *auth.SubjectPiiKeyStore,
	logger *slog.Logger,
) []eventHandlerFactory {
	return []eventHandlerFactory{
		{
			name: "password reset email",
			create: func() (eventHandler, error) {
				return auth.NewPasswordResetEmailToBeSentEventHandler(
					store,
					checkpointer,
					store,
					store,
					emailSender,
					cfg.AppURL,
					piiKeys,
					logger,
				)
			},
		},
		{
			name: "auth user projection",
			create: func() (eventHandler, error) {
				return auth.NewAuthUserProjectionEventHandler(
					store,
					checkpointer,
					store,
					authUsers,
					piiKeys,
					logger,
				)
			},
		},
		{
			name: "profile read model",
			create: func() (eventHandler, error) {
				return profileEvents.NewReadModelEventHandler(
					store,
					checkpointer,
					readModels.Profiles,
					bus,
					piiKeys,
					logger,
				)
			},
		},
		{
			name: "caseload students read model",
			create: func() (eventHandler, error) {
				return caseloadStudentsEvents.NewReadModelEventHandler(
					store,
					checkpointer,
					readModels.CaseloadStudents,
					bus,
					logger,
				)
			},
		},
		{
			name: "educator read model",
			create: func() (eventHandler, error) {
				return educatorEvents.NewReadModelEventHandler(
					store,
					checkpointer,
					readModels.Educators,
					bus,
					logger,
				)
			},
		},
		{
			name: "educator period read model",
			create: func() (eventHandler, error) {
				return educatorPeriodEvents.NewPeriodEducatorReadModelEventHandler(
					store,
					checkpointer,
					readModels.EducatorPeriods,
					bus,
					logger,
				)
			},
		},
		{
			name: "period read model",
			create: func() (eventHandler, error) {
				return periodEvents.NewPeriodReadModelEventHandler(
					store,
					checkpointer,
					readModels.Periods,
					bus,
					logger,
				)
			},
		},
		{
			name: "student read model",
			create: func() (eventHandler, error) {
				return studentEvents.NewStudentReadModelEventHandler(
					store,
					checkpointer,
					readModels.Students,
					bus,
					logger,
				)
			},
		},
		{
			name: "iep service read model",
			create: func() (eventHandler, error) {
				return iepServiceEvents.NewIEPServiceReadModelEventHandler(
					store,
					checkpointer,
					readModels.IEPServices,
					bus,
					logger,
				)
			},
		},
		{
			name: "period student read model",
			create: func() (eventHandler, error) {
				return periodStudentEvents.NewPeriodStudentReadModelEventHandler(
					store,
					checkpointer,
					readModels.StudentPeriods,
					bus,
					logger,
				)
			},
		},
	}
}

func StartEventHandlers(ctx context.Context, factories []eventHandlerFactory) ([]eventHandler, error) {
	handlers := make([]eventHandler, 0, len(factories))
	for _, factory := range factories {
		handler, err := factory.create()
		if err != nil {
			StopEventHandlers(handlers)
			return nil, fmt.Errorf("create %s event handler: %w", factory.name, err)
		}
		if err := handler.StartSubscribing(ctx); err != nil {
			StopEventHandlers(append(handlers, handler))
			return nil, fmt.Errorf("start %s event handler: %w", factory.name, err)
		}
		handlers = append(handlers, handler)
	}
	return handlers, nil
}

func StopEventHandlers(handlers []eventHandler) {
	for i := len(handlers) - 1; i >= 0; i-- {
		handlers[i].StopSubscribing()
	}
}
