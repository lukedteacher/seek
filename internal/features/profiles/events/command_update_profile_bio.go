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

type UpdateProfileBioCommand struct {
	User     models.User
	Bio      string
	Metadata eventstore.CommandMetadata
}

func UpdateProfileBioCommandHandler(ctx context.Context, command UpdateProfileBioCommand, saver eventstore.Saver, retriever eventstore.Retriever, keys auth.SubjectPiiKeyPort) error {
	if err := commandlimits.Assert(command); err != nil {
		return err
	}
	model, err := loadUpdateProfileBioContext(ctx, command, retriever, keys)
	if err != nil {
		return err
	}
	println("m: ", model.bio)
	println("n: ", model.nextBio)
	if model.bio == model.nextBio {
		return nil
	}
	event := NewProfileBioUpdatedEvent(model.eventID, model.nextBio, time.Now(), command.User.UserRegisteredID, model.subjectKey, nil)
	_, err = eventstore.SaveCommandEvents(ctx, saver, command.Metadata, []eventstore.DomainEvent{event}, model.position, model.events, model.query)
	return err
}

type updateProfileBioContext struct {
	userExists bool
	bio        string
	nextBio    string
	subjectKey protectedpii.SubjectDataKey
	eventID    string
	position   eventstore.Position
	events     []eventstore.ResolvedEvent
	query      eventstore.Query
}

func loadUpdateProfileBioContext(
	ctx context.Context,
	command UpdateProfileBioCommand,
	retriever eventstore.Retriever,
	keys auth.SubjectPiiKeyPort,
) (
	*updateProfileBioContext,
	error,
) {
	bio := strings.TrimSpace(command.Bio)
	if len(bio) > 280 {
		return nil, errors.New("bio must be 280 characters or fewer")
	}
	userQuery := registeredUserQuery(command.User.UserRegisteredID)
	bioQuery := profileUserEventQuery(ProfileBioUpdated, command.User.UserRegisteredID)
	query := combineQueries(userQuery, bioQuery)
	for _, q := range query.Criteria {
		for _, t := range q.Tags {
			println("k:", t.Key, "v:", t.Value)
		}
	}
	subjectKey, ok, err := keys.GetSubjectDataKey(ctx, command.User.UserRegisteredID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, eventstore.ErrSubjectKeyNotFound
	}
	model := &updateProfileBioContext{
		nextBio:    bio,
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

func (m *updateProfileBioContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case auth.UserRegistered:
		println("Hello")
		m.userExists = true
	case ProfileBioUpdated:
		println("world")
		m.bio = protectedpii.MustDecryptEventStringWithDataKey(protectedpii.FromEnv(), m.subjectKey, resolved.Event.Data, ProfileBioUpdatedBioField)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
