package httpapi

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/domain"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type loginResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

const sessionCookieName = "polycloud_session"

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, a.log, err)
		return
	}

	expectedEmail := a.cfg.AdminEmail
	expectedPass := a.cfg.AdminPassword

	emailMatch := subtle.ConstantTimeCompare([]byte(req.Email), []byte(expectedEmail)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(req.Password), []byte(expectedPass)) == 1

	if !emailMatch || !passMatch {
		writeError(w, a.log, domain.ErrUnauthorized)
		return
	}

	duration := 7 * 24 * time.Hour
	token, err := auth.CreateSessionToken(a.cfg.SessionSecret[:], a.cfg.DefaultUserID, req.Email, duration)
	if err != nil {
		writeError(w, a.log, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(duration),
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User: authUserResponse{
			ID:    a.cfg.DefaultUserID,
			Email: req.Email,
		},
	})
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, authUserResponse{
		ID:    a.userID(r),
		Email: a.cfg.AdminEmail,
	})
}
