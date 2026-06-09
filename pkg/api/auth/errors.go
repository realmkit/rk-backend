package auth

import "errors"

// ErrInvalidToken reports invalid bearer token validation.
var ErrInvalidToken = errors.New("invalid token")

// ErrDisabledUser reports that a local user cannot authenticate.
var ErrDisabledUser = errors.New("disabled user")
