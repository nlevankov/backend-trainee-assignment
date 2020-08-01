package models

import (
	"strings"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}
func (e modelError) Public() string {
	return strings.Replace(string(e), "models: ", "", 1)
}

const (
	ErrNoSuchEndpointExists modelError = "models: no such endpoint exists"
	ErrNoSuchHTTPMethod     modelError = "models: wrong http method"
)
