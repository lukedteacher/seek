package events

import (
	"context"
	"fmt"

	"seek/internal/appdb"
	"seek/internal/auth"
	"seek/internal/dbsql"
	"seek/internal/eventstore"
	"seek/internal/features/users/models"
	"seek/internal/protectedpii"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) GetUserProfileByID(ctx context.Context, userID string) (models.User, error) {
	var row *dbsql.GetUserProfileByUserIdRes
	err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetUserProfileByUserId(conn, userID)
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
		Avatar:           row.Avatar,
	}, nil
}

func (m *ReadModel) CreateProfileForUser(ctx context.Context, resolved eventstore.ResolvedEvent, keys auth.SubjectPiiKeyPort) error {
	data := resolved.Event.Data
	userRegisteredID, _ := data[auth.UserRegisteredEventID].(string)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateProfileForUser(conn, dbsql.CreateProfileForUserParams{
			UserId:                   userRegisteredID,
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateAvatar(
	ctx context.Context,
	resolved eventstore.ResolvedEvent,
) error {
	userRegisteredID, _ := eventstore.Scope(resolved.Event.Data)[auth.UserRegisteredEventID].(string)
	avatar, ok := resolved.Event.Data["avatar"].(string)
	if !ok {
		return fmt.Errorf("no avatar data in event")
	}
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceProfileUpdateAvatar(conn, dbsql.ProfileUpdateAvatarParams{
			UserId:                   userRegisteredID,
			Avatar:                   stringPtr(avatar),
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateBio(ctx context.Context, resolved eventstore.ResolvedEvent, keys auth.SubjectPiiKeyPort) error {
	userRegisteredID, _ := eventstore.Scope(resolved.Event.Data)[auth.UserRegisteredEventID].(string)
	subjectKey, ok, err := keys.GetSubjectDataKey(ctx, userRegisteredID)
	if err != nil {
		return err
	}
	if !ok {
		return eventstore.ErrSubjectKeyNotFound
	}
	bio := protectedpii.MustDecryptEventStringWithDataKey(protectedpii.FromEnv(), subjectKey, resolved.Event.Data, "bio")
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceProfileUpdateBio(conn, dbsql.ProfileUpdateBioParams{
			UserId:                   userRegisteredID,
			Bio:                      stringPtr(bio),
			LastEventCommitPosition:  resolved.Position.Commit,
			LastEventPreparePosition: resolved.Position.Prepare,
		})
	})
}

func stringPtr(value string) *string {
	return &value
}
