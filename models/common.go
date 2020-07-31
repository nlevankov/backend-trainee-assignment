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
