package main

import (
	"context"
	"fmt"
	"log/slog"
	"seek/internal/config"
	"seek/internal/eventstore"
	"seek/internal/features/period"
	"seek/internal/features/schedule"
	"seek/internal/features/student"
	"seek/internal/features/teacher"
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

func eventHandlerFactories(store *eventstore.EmbeddedOrisun, bus *natsbus.Bus, cfg config.Config, components appComponents, logger *slog.Logger) []eventHandlerFactory {
	return []eventHandlerFactory{
		{
			name: "period read model",
			create: func() (eventHandler, error) {
				return period.NewPeriodReadModelEventHandler(store, components.checkpointer, components.periodReadModel, bus, logger)
			},
		},
		{
			name: "schedule read model",
			create: func() (eventHandler, error) {
				return schedule.NewScheduleReadModelEventHandler(store, components.checkpointer, components.scheduleReadModel, bus, logger)
			},
		},
		{
			name: "student read model",
			create: func() (eventHandler, error) {
				return student.NewStudentReadModelEventHandler(store, components.checkpointer, components.studentReadModel, bus, logger)
			},
		},
		{
			name: "teacher read model",
			create: func() (eventHandler, error) {
				return teacher.NewTeacherReadModelEventHandler(store, components.checkpointer, components.teacherReadModel, bus, logger)
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