// ValidateUsername is intentionally NOT applied on the OAuth/SSO path —
// external IdPs (GitHub, GitLab, Google, OIDC, SAML, LDAP) emit usernames
// containing Unicode/dots that we cannot reject without breaking SSO. The DB
// schema VARCHAR(255) accommodates them; the slug-derivation layer
// (service/organization/service_personal.go) sanitizes Unicode usernames into
// `user-<id>-workspace` fallback.
package user

import (
	"errors"
	"regexp"
)

const (
	UsernameMinLen = 3
	UsernameMaxLen = 50
)

var ErrInvalidUsername = errors.New("username must contain only letters, numbers, hyphens, and underscores, length 3-50")

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidateUsername(s string) error {
	if len(s) < UsernameMinLen || len(s) > UsernameMaxLen {
		return ErrInvalidUsername
	}
	if !usernamePattern.MatchString(s) {
		return ErrInvalidUsername
	}
	return nil
}
