package proxy

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

const basicAuthRealm = `Basic realm="portal"`

// Authorizer determines whether an incoming request may reach a backend.
// Implementations can support additional providers without changing Server.
type Authorizer interface {
	Authorize(*http.Request) bool
}

type basicAuthorizer struct {
	username string
	password string
}

// NewBasicAuthorizer returns an HTTP Basic authorization provider.
func NewBasicAuthorizer(username, password string) Authorizer {
	return basicAuthorizer{username: username, password: password}
}

func (a basicAuthorizer) Authorize(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}

	expectedUsername := sha256.Sum256([]byte(a.username))
	expectedPassword := sha256.Sum256([]byte(a.password))
	providedUsername := sha256.Sum256([]byte(username))
	providedPassword := sha256.Sum256([]byte(password))
	usernameMatches := subtle.ConstantTimeCompare(providedUsername[:], expectedUsername[:]) == 1
	passwordMatches := subtle.ConstantTimeCompare(providedPassword[:], expectedPassword[:]) == 1
	return usernameMatches && passwordMatches
}
