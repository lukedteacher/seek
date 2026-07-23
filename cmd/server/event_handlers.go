package main

import (
	"context"
	"fmt"
	"log/slog"
	"seek/internal/auth"
	"seek/internal/config"
	"seek/internal/eventstore"
	educatorEvents "seek/internal/features/educators/events"
	iepServiceEvents "seek/internal/features/iep_services/events"
	periodEvents "seek/internal/features/periods/events"
	periodScheduleEvents "seek/internal/features/periods_schedules/events"
	periodStudentEvents "seek/internal/features/periods_students/events"
	profileEvents "seek/internal/features/profiles/events"
	scheduleEvents "seek/internal/features/schedules/events"
	studentEvents "seek/internal/features/students/events"
	teacherEvents "seek/internal/features/teachers/events"
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
					components.profileReadModel,
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
					components.educatorReadModel,
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
					components.periodReadModel,
					bus,
					logger,
				)
			},
		},
		{
			name: "schedule read model",
			create: func() (eventHandler, error) {
				return scheduleEvents.NewScheduleReadModelEventHandler(
					store,
					components.checkpointer,
					components.scheduleReadModel,
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
					components.studentReadModel,
					bus,
					logger,
				)
			},
		},
		{
			name: "teacher read model",
			create: func() (eventHandler, error) {
				return teacherEvents.NewTeacherReadModelEventHandler(
					store,
					components.checkpointer,
					components.teacherReadModel,
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
					components.iepServiceReadModel,
					bus,
					logger,
				)
			},
		},
		{
			name: "period schedule read model",
			create: func() (eventHandler, error) {
				return periodScheduleEvents.NewPeriodScheduleReadModelEventHandler(
					store,
					components.checkpointer,
					components.periodScheduleReadModel,
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
					components.periodStudentReadModel,
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
