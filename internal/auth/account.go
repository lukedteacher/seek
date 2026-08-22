package auth

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/users/models"
)

func userFromSessionRow(row *dbsql.UserBySessionTokenRes) (models.User, error) {
	if row == nil {
		return models.User{}, appdb.ErrNoRows
	}
	return models.User{
		ID:               row.Id,
		UserRegisteredID: row.UserRegisteredId,
		Email:            row.Email,
		Username:         row.Username,
		Avatar:           row.Avatar,
		Bio:              row.Bio,
	}, nil
}

func userFromRegisteredRow(row *dbsql.UserByRegisteredIdRes) (models.User, error) {
	if row == nil {
		return models.User{}, appdb.ErrNoRows
	}
	return models.User{
		ID:               row.Id,
		UserRegisteredID: row.UserRegisteredId,
		Email:            row.Email,
		Username:         row.Username,
		Avatar:           row.Avatar,
		Bio:              row.Bio,
	}, nil
}

func userFromIDOrRegisteredRow(row *dbsql.UserByIdorRegisteredIdRes) (models.User, error) {
	if row == nil {
		return models.User{}, appdb.ErrNoRows
	}
	return models.User{
		ID:               row.Id,
		UserRegisteredID: row.UserRegisteredId,
		Email:            row.Email,
		Username:         row.Username,
		Avatar:           row.Avatar,
		Bio:              row.Bio,
	}, nil
}

func stringPtr(value string) *string {
	return &value
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func numericCode(size int) (string, error) {
	var b strings.Builder
	for i := 0; i < size; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}
