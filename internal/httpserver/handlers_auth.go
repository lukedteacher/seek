package httpserver

import (
	"fmt"
	"net/http"

	"seek/internal/auth"
	"seek/internal/eventstore"
	"seek/internal/ui/core/corepages"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) authRoutes(r chi.Router) {
	r.With(noCache).Get("/register", render(corepages.Register(nil)))
	r.With(noCache, signupRateLimit).Post("/register", s.register)
	r.With(noCache).Get("/login", render(corepages.Login(nil)))
	r.With(noCache, loginRateLimit).Post("/login", s.login)
	r.Post("/logout", s.logout)
	r.With(noCache).Get("/forgot-password", render(corepages.ForgotPassword(nil)))
	r.With(noCache, forgotPasswordRateLimit).Post("/forgot-password", s.forgotPassword)
	r.With(noCache).Get("/reset-password/{token}", func(w http.ResponseWriter, r *http.Request) {
		_ = corepages.ResetPassword(chi.URLParam(r, "token"), nil).Render(r.Context(), w)
	})
	r.With(noCache, resetPasswordRateLimit).Post("/reset-password/{token}", s.resetPassword)
	r.With(noCache).Get("/register/{userID}/validate-email", func(w http.ResponseWriter, r *http.Request) {
		_ = corepages.ValidateEmail(chi.URLParam(r, "userID"), nil).Render(r.Context(), w)
	})
}

func (s Server) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setRequestAction(r, "auth.register", nil)
	if err := r.ParseForm(); err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error { return flashError(sse, err.Error()) })
		return
	}
	_, err := auth.RegisterUserCommandHandler(ctx, auth.RegisterUserCommand{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Metadata: eventstore.HTTPCommandMetadata(r, ""),
	}, s.EventSaver, s.EventRetriever, s.PIIKeys)
	if err != nil {
		patchTempl(w, r, corepages.RegisterForm(map[string]string{"error": err.Error()}), datastar.WithSelectorID("auth-page"))
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/login")
}

func (s Server) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setRequestAction(r, "auth.login", nil)
	if err := r.ParseForm(); err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error { return flashError(sse, err.Error()) })
		return
	}
	user, token, err := s.Sessions.Login(ctx, r.FormValue("email"), r.FormValue("password"))
	userRegisteredID := ""
	if err == nil {
		userRegisteredID = user.UserRegisteredID
	}
	_ = auth.RecordLoginAttemptCommandHandler(ctx, auth.RecordLoginAttemptCommand{
		AttemptedIdentifier: r.FormValue("email"),
		IPAddress:           r.RemoteAddr,
		UserRegisteredID:    userRegisteredID,
		Succeeded:           err == nil,
		Metadata:            eventstore.HTTPCommandMetadata(r, userRegisteredID),
	}, s.EventSaver)
	if err != nil {
		patchTempl(w, r, corepages.LoginForm(map[string]string{"error": err.Error()}), datastar.WithSelectorID("auth-page"))
		return
	}
	s.Sessions.SetSessionCookie(w, token)
	path := "/"
	if !user.EmailVerified {
		path = fmt.Sprintf("/register/%s/validate-email", user.ID)
	}
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error { return sse.Redirect(path) })
}

func (s Server) logout(w http.ResponseWriter, r *http.Request) {
	setRequestAction(r, "auth.logout", nil)
	if cookie, err := r.Cookie(s.Sessions.SessionCookieName()); err == nil {
		_ = s.Sessions.Logout(r.Context(), cookie.Value)
	}
	s.Sessions.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s Server) forgotPassword(w http.ResponseWriter, r *http.Request) {
	setRequestAction(r, "auth.forgot_password", nil)
	_ = r.ParseForm()
	_, _ = auth.RequestPasswordResetCommandHandler(r.Context(), auth.RequestPasswordResetCommand{
		EmailAddress: r.FormValue("email"),
		Metadata:     eventstore.HTTPCommandMetadata(r, ""),
	}, s.PasswordCredentials, s.EventSaver, s.EventRetriever, s.PIIKeys)
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error { return sse.Redirect("/login") })
}

func (s Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	setRequestAction(r, "auth.reset_password", nil)
	_ = r.ParseForm()
	token := chi.URLParam(r, "token")
	if err := auth.ResetPasswordCommandHandler(r.Context(), auth.ResetPasswordCommand{
		Token:    token,
		Password: r.FormValue("password"),
		Metadata: eventstore.HTTPCommandMetadata(r, ""),
	}, s.Verifications, s.AuthUsers, s.EventSaver, s.EventRetriever); err != nil {
		patchTempl(w, r, corepages.ResetPasswordForm(token, map[string]string{"error": err.Error()}), datastar.WithSelectorID("auth-page"))
		return
	}
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error { return sse.Redirect("/login") })
}
