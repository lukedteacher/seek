package events

import (
	"context"

	"seek/internal/appdb"
	"seek/internal/auth"
	"seek/internal/dbsql"
	"seek/internal/eventstore"
	"seek/internal/features/users/models"
	"seek/internal/protectedpii"

	"zombiezen.com/go/sqlite"
)

const (
	userRegistered = "UserRegistered"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) User(ctx context.Context, userRegisteredID string) (models.User, error) {
	var row *dbsql.ProfileUserRes
	err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceProfileUser(conn, userRegisteredID)
		return err
	})
	if err != nil {
		return models.User{}, err
	}
	if row == nil {
		return models.User{}, appdb.ErrNoRows
	}
	return models.User{
		ID:               row.UserId,
		UserRegisteredID: row.UserId,
		Email:            row.Email,
		Image:            row.Image,
		Bio:              row.Bio,
		HeaderImageURL:   row.HeaderImageUrl,
	}, nil
}

func (m *ReadModel) UpsertRegisteredUser(ctx context.Context, resolved eventstore.ResolvedEvent, keys auth.SubjectPiiKeyPort) error {
	data := resolved.Event.Data
	userRegisteredID, _ := data[auth.FieldUserRegisteredID].(string)
	protector := protectedpii.FromEnv()
	subjectKey, ok, err := keys.GetSubjectDataKey(ctx, userRegisteredID)
	if err != nil {
		return err
	}
	if !ok {
		return eventstore.ErrSubjectKeyNotFound
	}
	emailAddress := protectedpii.MustDecryptEventStringWithDataKey(protector, subjectKey, data, "email")
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpsertRegisteredProfileUser(conn, dbsql.UpsertRegisteredProfileUserParams{
			UserId:                   userRegisteredID,
			Email:                    stringPtr(emailAddress),
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateBio(ctx context.Context, resolved eventstore.ResolvedEvent, keys auth.SubjectPiiKeyPort) error {
	userRegisteredID, _ := eventstore.Scope(resolved.Event.Data)[auth.FieldUserRegisteredID].(string)
	println("u: ", userRegisteredID)
	subjectKey, ok, err := keys.GetSubjectDataKey(ctx, userRegisteredID)
	println("e: ", ok)
	if err != nil {
		return err
	}
	if !ok {
		return eventstore.ErrSubjectKeyNotFound
	}
	bio := protectedpii.MustDecryptEventStringWithDataKey(protectedpii.FromEnv(), subjectKey, resolved.Event.Data, "bio")
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpsertProfileBio(conn, dbsql.UpsertProfileBioParams{
			UserId:                   userRegisteredID,
			Bio:                      stringPtr(bio),
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateImage(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	userRegisteredID, _ := eventstore.Scope(resolved.Event.Data)[auth.FieldUserRegisteredID].(string)
	url, _ := resolved.Event.Data[ProfileImageURLField].(string)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpsertProfileImage(conn, dbsql.UpsertProfileImageParams{
			UserId:                   userRegisteredID,
			Image:                    stringPtr(url),
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateHeaderImage(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	userRegisteredID, _ := eventstore.Scope(resolved.Event.Data)[auth.FieldUserRegisteredID].(string)
	url, _ := resolved.Event.Data[ProfileImageURLField].(string)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpsertProfileHeaderImage(conn, dbsql.UpsertProfileHeaderImageParams{
			UserId:                   userRegisteredID,
			HeaderImageUrl:           stringPtr(url),
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func stringPtr(value string) *string {
	return &value
}
