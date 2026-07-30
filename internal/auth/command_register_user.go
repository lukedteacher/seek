package auth

import (
	"context"
	"errors"
	"strings"

	"seek/internal/commandlimits"
	"seek/internal/eventstore"
	"seek/internal/protectedpii"
	"seek/pkg/uuidv7"

	"golang.org/x/crypto/bcrypt"
)

type RegisterUserCommand struct {
	Email    string
	Password string
	Metadata eventstore.CommandMetadata // changed this from orisun
}

type RegisterUserResult struct {
	EventID      string
	Email        string
	PasswordHash string
}

func RegisterUserCommandHandler(
	ctx context.Context,
	command RegisterUserCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	keys SubjectPiiKeyPort,
) (
	RegisterUserResult,
	error,
) {
	if err := commandlimits.Assert(command); err != nil {
		return RegisterUserResult{}, err
	}
	if len(command.Password) < 6 {
		return RegisterUserResult{}, errors.New("invalid registration input")
	}
	model, err := loadRegisterUserContext(ctx, command, retriever)
	if err != nil {
		return RegisterUserResult{}, err
	}
	if model.existingEmail {
		return RegisterUserResult{}, errors.New("user already exists")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterUserResult{}, err
	}

	subjectKey, err := keys.GetOrCreateSubjectDataKey(ctx, model.id)
	if err != nil {
		return RegisterUserResult{}, err
	}
	event := NewUserRegisteredEvent(
		model.id,
		model.email,
		string(passwordHash),
		subjectKey,
		nil,
	)
	if _, err := eventstore.SaveCommandEvents(
		ctx,
		saver,
		command.Metadata,
		[]eventstore.DomainEvent{event},
		model.position,
		model.events,
		model.query,
	); err != nil {
		return RegisterUserResult{}, err
	}
	return RegisterUserResult{
		EventID:      model.id,
		Email:        model.email,
		PasswordHash: string(passwordHash),
	}, nil
}

type registerUserContext struct {
	existingEmail bool
	id            string
	email         string
	emailHash     string
	position      eventstore.Position
	events        []eventstore.ResolvedEvent
	query         eventstore.Query
}

func loadRegisterUserContext(
	ctx context.Context,
	command RegisterUserCommand,
	retriever eventstore.Retriever,
) (
	*registerUserContext,
	error,
) {
	email := strings.ToLower(strings.TrimSpace(command.Email))
	if email == "" {
		return nil, errors.New("invalid registration input")
	}

	protector := protectedpii.FromEnv()
	emailHash := protector.BlindIndex(UserRegisteredEmailField, email)
	query := userRegisteredByEmailQuery(emailHash)
	latest, err := retriever.GetLatestByCriteria(ctx, query.Criteria)
	if err != nil {
		return nil, err
	}
	events := eventstore.EventsFromLatest(latest.Results)

	model := &registerUserContext{
		id:        uuidv7.NewString(),
		email:     email,
		emailHash: emailHash,
		position:  latest.ContextPosition,
		events:    events,
		query:     query,
	}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *registerUserContext) handle(resolved eventstore.ResolvedEvent) {
	if resolved.Event.EventType == UserRegistered {
		emailHash, _ := resolved.Event.Data[UserRegisteredEmailHashField].(string)
		if emailHash == m.emailHash {
			m.existingEmail = true
		}
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
