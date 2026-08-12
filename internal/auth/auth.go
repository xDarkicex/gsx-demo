// Package auth provides password hashing, session management, and
// the nanite middleware that guards the dashboard.
package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/xDarkicex/nanite"
	"golang.org/x/crypto/bcrypt"

	"github.com/xDarkicex/gsx-demo/models"
)

// Store is the durable session backend. The db package implements
// it over libraVDB's relational surface, so sessions survive
// server restarts.
type Store interface {
	CreateSession(ctx context.Context, token, userID string, expiresAt time.Time) error
	SessionUser(ctx context.Context, token string) (*models.User, error)
	DeleteSession(ctx context.Context, token string) error
}

// Default is the wired session store. main sets it to db.Default
// after opening the database; handlers and @action bodies share it.
var Default Store

// SetStore wires the session backend (called by main at boot).
func SetStore(s Store) { Default = s }

// SessionCookie is the HttpOnly cookie that carries the session id.
const SessionCookie = "gsx_session"

// SessionTTL is how long a session lives before expiring.
const SessionTTL = 7 * 24 * time.Hour

// Create issues a new session token and persists it durably.
func Create(u *models.User) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	tok := hex.EncodeToString(buf)
	if Default == nil {
		return "", errors.New("auth: no session store wired")
	}
	if err := Default.CreateSession(context.Background(), tok, u.ID,
		time.Now().Add(SessionTTL)); err != nil {
		return "", err
	}
	return tok, nil
}

// Get resolves a token to its user via the durable store, or nil
// if missing/expired.
func Get(tok string) *models.User {
	if Default == nil || tok == "" {
		return nil
	}
	u, err := Default.SessionUser(context.Background(), tok)
	if err != nil || u == nil {
		return nil
	}
	return u
}

// Delete removes a session (logout).
func Delete(tok string) {
	if Default == nil {
		return
	}
	_ = Default.DeleteSession(context.Background(), tok)
}

// HashPassword bcrypt-hashes a plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword reports whether plain matches the stored hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// RequireUser is nanite middleware: resolve the session cookie to a
// user (via the Default session store), stash it on the context,
// or redirect to /login.
func RequireUser(c *nanite.Context, next func()) {
	cookie, err := c.Request.Cookie(SessionCookie)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	u := Get(cookie.Value)
	if u == nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	c.Set("user", u)
	next()
}

// CurrentUser returns the authenticated user from a handler context.
func CurrentUser(c *nanite.Context) *models.User {
	if v, ok := c.Get("user"); ok {
		if u, ok := v.(*models.User); ok {
			return u
		}
	}
	return nil
}

// UserFromRequest resolves the session cookie to a user directly
// from an *http.Request — used inside @action bodies, where there
// is no nanite context.
func UserFromRequest(r *http.Request) *models.User {
	if r == nil {
		return nil
	}
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return nil
	}
	return Get(cookie.Value)
}

// AttachUser resolves an optional session for public pages — it
// applies when a valid cookie exists and never redirects. Guards
// on protected routes still use RequireUser.
func AttachUser(c *nanite.Context, next func()) {
	if cookie, err := c.Request.Cookie(SessionCookie); err == nil {
		if u := Get(cookie.Value); u != nil {
			c.Set("user", u)
		}
	}
	next()
}
