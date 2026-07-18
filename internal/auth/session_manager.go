package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"seek/internal/appdb"
	"seek/internal/commandlimits"
	"seek/internal/dbsql"
	"seek/internal/domain/models"
	"seek/internal/uuidv7"

	"golang.org/x/crypto/bcrypt"
	"zombiezen.com/go/sqlite"
)

type SessionManager struct {
	db            *appdb.DB
	users         *AuthUserStore
	secureCookie  bool
	sessionCookie string
}

func NewSessionManager(db *appdb.DB, users *AuthUserStore, secureCookie bool) *SessionManager {
	return &SessionManager{
		db:            db,
		users:         users,
		secureCookie:  secureCookie,
		sessionCookie: "seek-session",
	}
}

func (s *SessionManager) Login(ctx context.Context, email, password string) (models.User, string, error) {
	// checks the email and password against length constrains
	if err := commandlimits.Assert(struct {
		Email    string
		Password string
	}{Email: email, Password: password}); err != nil {
		return models.User{}, "", err
	}
	// checks if the user exists, returns the password hash
	user, hash, err := s.users.UserByEmailWithPassword(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return models.User{}, "", errors.New("invalid email or password")
	}
	// hashes the provied password and compares
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return models.User{}, "", errors.New("invalid email or password")
	}
	token, err := randomToken(32)
	if err != nil {
		return models.User{}, "", err
	}
	err = s.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateAuthSession(conn, dbsql.CreateAuthSessionParams{
			Id:        uuidv7.NewString(),
			Token:     token,
			UserId:    user.ID,
			ExpiresAt: appdb.SQLTime(time.Now().Add(90 * 24 * time.Hour)),
		})
	})
	return user, token, err
}

func (s *SessionManager) Logout(ctx context.Context, token string) error {
	return s.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteAuthSessionByToken(conn, token)
	})
}

func (s *SessionManager) CurrentUser(ctx context.Context, r *http.Request) (models.User, bool, error) {
	cookie, err := r.Cookie(s.sessionCookie)
	if err != nil || cookie.Value == "" {
		return models.User{}, false, nil
	}
	user, err := s.users.UserBySessionToken(ctx, cookie.Value)
	if err != nil {
		if errors.Is(err, appdb.ErrNoRows) {
			return models.User{}, false, nil
		}
		return models.User{}, false, err
	}
	return user, true, nil
}

func (s *SessionManager) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: s.sessionCookie, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.secureCookie, Expires: time.Now().Add(90 * 24 * time.Hour)})
}

func (s *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: s.sessionCookie, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.secureCookie, MaxAge: -1})
}

func (s *SessionManager) SessionCookieName() string {
	return s.sessionCookie
}
