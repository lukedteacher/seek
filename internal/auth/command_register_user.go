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
	Username    string
	Email       string
	Password    string
	GivenName   string
	FamilyName  string
	YearOfBirth int
	Metadata    eventstore.CommandMetadata // changed this from orisun
}

type RegisterUserResult struct {
	UserRegisteredID string
	Username         string
	Email            string
	GivenName        string
	FamilyName       string
	PasswordHash     string
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
	if model.existingUsername || model.existingEmail {
		return RegisterUserResult{}, errors.New("user already exists")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterUserResult{}, err
	}

	subjectKey, err := keys.GetOrCreateSubjectDataKey(ctx, model.userRegisteredID)
	if err != nil {
		return RegisterUserResult{}, err
	}
	event := NewUserRegisteredEvent(
		model.userRegisteredID,
		model.username,
		model.email,
		model.givenName,
		model.familyName,
		command.YearOfBirth,
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
		UserRegisteredID: model.userRegisteredID,
		Username:         model.username,
		Email:            model.email,
		GivenName:        model.givenName,
		FamilyName:       model.familyName,
		PasswordHash:     string(passwordHash),
	}, nil
}

type registerUserContext struct {
	existingUsername bool
	existingEmail    bool
	userRegisteredID string
	username         string
	usernameHash     string
	email            string
	emailHash        string
	givenName        string
	familyName       string
	position         eventstore.Position
	events           []eventstore.ResolvedEvent
	query            eventstore.Query
}

func loadRegisterUserContext(ctx context.Context, command RegisterUserCommand, retriever eventstore.Retriever) (*registerUserContext, error) {
	username := strings.TrimSpace(command.Username)
	email := strings.ToLower(strings.TrimSpace(command.Email))
	if len(username) < 4 || email == "" {
		return nil, errors.New("invalid registration input")
	}

	protector := protectedpii.FromEnv()
	usernameHash := protector.BlindIndex(UserRegisteredUsernameField, username)
	emailHash := protector.BlindIndex(UserRegisteredEmailField, email)
	query := userRegisteredByUsernameOrEmailQuery(usernameHash, emailHash)
	latest, err := retriever.GetLatestByCriteria(ctx, query.Criteria)
	if err != nil {
		return nil, err
	}
	events := eventstore.EventsFromLatest(latest.Results)

	model := &registerUserContext{
		userRegisteredID: uuidv7.NewString(),
		username:         username,
		usernameHash:     usernameHash,
		email:            email,
		emailHash:        emailHash,
		givenName:        strings.TrimSpace(command.GivenName),
		familyName:       strings.TrimSpace(command.FamilyName),
		position:         latest.ContextPosition,
		events:           events,
		query:            query,
	}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *registerUserContext) handle(resolved eventstore.ResolvedEvent) {
	if resolved.Event.EventType == UserRegistered {
		usernameHash, _ := resolved.Event.Data[UserRegisteredUsernameHashField].(string)
		emailHash, _ := resolved.Event.Data[UserRegisteredEmailHashField].(string)
		if usernameHash == m.usernameHash {
			m.existingUsername = true
		}
		if emailHash == m.emailHash {
			m.existingEmail = true
		}
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
