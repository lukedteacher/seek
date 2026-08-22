package events

import (
	"context"
	"errors"
	"strings"
	"time"

	"seek/internal/auth"
	"seek/internal/commandlimits"
	"seek/internal/eventstore"
	"seek/internal/features/users/models"
	"seek/internal/protectedpii"
	"seek/pkg/uuidv7"
)

type UpdateProfileAvatarCommand struct {
	User     models.User
	Avatar   string
	Metadata eventstore.CommandMetadata
}

func UpdateProfileAvatarCommandHandler(
	ctx context.Context,
	command UpdateProfileAvatarCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	keys auth.SubjectPiiKeyPort,
) error {
	if err := commandlimits.Assert(command); err != nil {
		return err
	}
	model, err := loadUpdateProfileAvatarContext(ctx, command, retriever, keys)
	if err != nil {
		return err
	}
	if model.avatar == model.nextAvatar {
		return nil
	}
	event := NewProfileAvatarUpdatedEvent(
		model.eventID,
		model.nextAvatar,
		time.Now(),
		command.User.UserRegisteredID,
		nil,
	)
	_, err = eventstore.SaveCommandEvents(
		ctx,
		saver,
		command.Metadata,
		[]eventstore.DomainEvent{event},
		model.position,
		model.events,
		model.query,
	)
	return err
}

type updateProfileAvatarContext struct {
	userExists bool
	avatar     string
	nextAvatar string
	subjectKey protectedpii.SubjectDataKey
	eventID    string
	position   eventstore.Position
	events     []eventstore.ResolvedEvent
	query      eventstore.Query
}

func loadUpdateProfileAvatarContext(
	ctx context.Context,
	command UpdateProfileAvatarCommand,
	retriever eventstore.Retriever,
	keys auth.SubjectPiiKeyPort,
) (
	*updateProfileAvatarContext,
	error,
) {
	avatar := strings.TrimSpace(command.Avatar)
	if len(avatar) > 280 {
		return nil, errors.New("avatar must be 280 characters or fewer")
	}
	userQuery := registeredUserQuery(command.User.UserRegisteredID)
	avatarQuery := profileUserEventQuery(ProfileAvatarUpdated, command.User.UserRegisteredID)
	query := combineQueries(userQuery, avatarQuery)
	subjectKey, ok, err := keys.GetSubjectDataKey(ctx, command.User.UserRegisteredID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, eventstore.ErrSubjectKeyNotFound
	}
	model := &updateProfileAvatarContext{
		nextAvatar: avatar,
		subjectKey: subjectKey,
		eventID:    uuidv7.NewString(),
		position:   eventstore.NoEventPosition,
		query:      query,
	}
	latest, err := retriever.GetLatestByCriteria(ctx, query.Criteria)
	if err != nil {
		return nil, err
	}
	model.events = eventstore.EventsFromLatest(latest.Results)
	model.position = latest.ContextPosition
	for _, event := range model.events {
		model.handle(event)
	}
	if !model.userExists {
		return nil, eventstore.ErrUserNotActive
	}
	return model, nil
}

func (m *updateProfileAvatarContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case auth.UserRegistered:
		m.userExists = true
	case ProfileAvatarUpdated:
		m.avatar = protectedpii.MustDecryptEventStringWithDataKey(
			protectedpii.FromEnv(),
			m.subjectKey,
			resolved.Event.Data,
			FieldProfileAvatarUpdatedAvatar,
		)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
