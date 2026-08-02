package main

import (
	"context"
	"fmt"
	"log/slog"
	"seek/internal/auth"
	"seek/internal/config"
	"seek/internal/eventstore"
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

func eventHandlerFactories(
	store *eventstore.EmbeddedOrisun,
	bus *natsbus.Bus,
	cfg config.Config,
	components appComponents,
	logger *slog.Logger,
) []eventHandlerFactory {
	return []eventHandlerFactory{
		{
			name: "registration OTP",
			create: func() (eventHandler, error) {
				return auth.NewRegistrationOTPToBeGeneratedEventHandler(
					store,
					components.checkpointer,
					store,
					store,
					logger,
				)
			},
		},
		{
			name: "email validation OTP",
			create: func() (eventHandler, error) {
				return auth.NewEmailValidationOTPToBeSentEventHandler(
					store,
					components.checkpointer,
					store,
					store,
					components.emailSender,
					components.piiKeys,
					logger,
				)
			},
		},
		{
			name: "password reset email",
			create: func() (eventHandler, error) {
				return auth.NewPasswordResetEmailToBeSentEventHandler(
					store,
					components.checkpointer,
					store,
					store,
					components.emailSender,
					cfg.AppURL,
					components.piiKeys,
					logger,
				)
			},
		},
		{
			name: "auth user projection",
			create: func() (eventHandler, error) {
				return auth.NewAuthUserProjectionEventHandler(
					store,
					components.checkpointer,
					store,
					components.authUsers,
					components.verifications,
					components.piiKeys,
					logger,
				)
			},
		},
		{
			name: "profile read model",
			create: func() (eventHandler, error) {
				return profileEvents.NewReadModelEventHandler(
					store,
					components.checkpointer,
					components.readModels.Profiles,
					bus,
					components.piiKeys,
					logger,
				)
			},
		},
		{
			name: "educator read model",
			create: func() (eventHandler, error) {
				return educatorEvents.NewReadModelEventHandler(
					store,
					components.checkpointer,
					components.readModels.Educators,
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
					components.checkpointer,
					components.readModels.EducatorPeriods,
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
					components.checkpointer,
					components.readModels.Periods,
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
					components.checkpointer,
					components.readModels.Students,
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
					components.checkpointer,
					components.readModels.IEPServices,
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
					components.checkpointer,
					components.readModels.StudentPeriods,
					bus,
					logger,
				)
			},
		},
	}
}

func startEventHandlers(ctx context.Context, factories []eventHandlerFactory) ([]eventHandler, error) {
	handlers := make([]eventHandler, 0, len(factories))
	for _, factory := range factories {
		handler, err := factory.create()
		if err != nil {
			stopEventHandlers(handlers)
			return nil, fmt.Errorf("create %s event handler: %w", factory.name, err)
		}
		if err := handler.StartSubscribing(ctx); err != nil {
			stopEventHandlers(append(handlers, handler))
			return nil, fmt.Errorf("start %s event handler: %w", factory.name, err)
		}
		handlers = append(handlers, handler)
	}
	return handlers, nil
}

func stopEventHandlers(handlers []eventHandler) {
	for i := len(handlers) - 1; i >= 0; i-- {
		handlers[i].StopSubscribing()
	}
}
