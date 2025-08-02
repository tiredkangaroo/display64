package spotify

import "errors"

var (
	ErrAuthorizationRequired = errors.New("authorization required")
)
