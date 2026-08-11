// Package auth provides password hashing, session management, and
// the nanite middleware that guards the dashboard.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/xDarkicex/nanite"
	"golang.org/x/crypto/bcrypt"

	"github.com/xDarkicex/gsx-demo/models"
)

// Default is the demo's session store. Handlers and @action bodies
// share this one instance.
var Default = NewSessions()

// SessionCookie is the HttpOnly cookie that carries the session id.
const SessionCookie = "gsx_session"

// SessionTTL is how long a session lives before expiring.
const SessionTTL = 7 * 24 * time.Hour

// Sessions is an in-memory session store: token → user.
// Fine for a demo; swap for a sessions table for real deployments.
type Sessions struct {
	mu sync.RWMutex
	m  map[string]session
}

type session struct {
	user *models.User
	exp  time.Time
}

// NewSessions returns an empty session store.
func NewSessions() *Sessions {
	return &Sessions{m: make(map[string]session)}
}

// Create issues a new session for the user and returns its token.
func (s *Sessions) Create(u *models.User) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	tok := hex.EncodeToString(buf)
	s.mu.Lock()
	s.m[tok] = session{user: u, exp: time.Now().Add(SessionTTL)}
	s.mu.Unlock()
	return tok, nil
}

// Get resolves a token to its user, or nil if missing/expired.
func (s *Sessions) Get(tok string) *models.User {
	s.mu.RLock()
	se, ok := s.m[tok]
	s.mu.RUnlock()
	if !ok || time.Now().After(se.exp) {
		return nil
	}
	return se.user
}

// Delete removes a session (logout).
func (s *Sessions) Delete(tok string) {
	s.mu.Lock()
	delete(s.m, tok)
	s.mu.Unlock()
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
	u := Default.Get(cookie.Value)
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
	return Default.Get(cookie.Value)
}
