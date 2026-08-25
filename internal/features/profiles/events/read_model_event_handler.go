package events

import (
	"context"
	"fmt"
	"log/slog"

	"seek/internal/auth"
	"seek/internal/eventstore"
)

const ProfileReadModelEventHandlerName = "profile_read_model_event_handler"

type ReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel *ReadModel
	publisher eventstore.Publisher
	keys      auth.SubjectPiiKeyPort
}

func NewReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel *ReadModel,
	publisher eventstore.Publisher,
	keys auth.SubjectPiiKeyPort,
	logger *slog.Logger,
) (
	*ReadModelEventHandler,
	error,
) {
	handler := &ReadModelEventHandler{readModel: readModel, publisher: publisher, keys: keys}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            ProfileReadModelEventHandlerName,
		Query:           readModelEventHandlerQuery(),
		Logger:          logger,
		MaxEventRetries: -1,
		HandleEvent:     handler.handle,
	})
	if err != nil {
		return nil, err
	}
	handler.global = global
	return handler, nil
}

func (h *ReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *ReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func (h *ReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	var userRegisteredID string
	switch resolved.Event.EventType {
	case auth.UserRegistered:
		userRegisteredID, _ = resolved.Event.Data[auth.UserRegisteredEventID].(string)
		if err := h.readModel.CreateProfileForUser(ctx, resolved, h.keys); err != nil {
			return err
		}
	case ProfileAvatarUpdated:
		userRegisteredID, _ = eventstore.Scope(resolved.Event.Data)[auth.UserRegisteredEventID].(string)
		if err := h.readModel.UpdateAvatar(ctx, resolved); err != nil {
			return err
		}
	case ProfileBioUpdated:
		userRegisteredID, _ = eventstore.Scope(resolved.Event.Data)[auth.UserRegisteredEventID].(string)
		if err := h.readModel.UpdateBio(ctx, resolved, h.keys); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled profile read model event type %q", resolved.Event.EventType)
	}
	if userRegisteredID == "" {
		return nil
	}
	return h.publisher.Publish(
		ctx, Channel(userRegisteredID),
		map[string]string{"user_registered_event_id": userRegisteredID},
	)
}

func readModelEventHandlerQuery() eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: auth.UserRegisteredEventID},
			}},
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ProfileAvatarUpdated.String()},
			}},
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ProfileBioUpdated.String()},
			}},
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ProfileImageUploaded.String()},
			}},
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ProfileHeaderImageUploaded.String()},
			}},
		},
	}
}
