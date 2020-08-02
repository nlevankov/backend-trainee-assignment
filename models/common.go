package models

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	return e.Error()
}

const (
	ErrNoSuchEndpointExists modelError = "No such endpoint exists"
	ErrNoSuchHTTPMethod     modelError = "Wrong http method"
)
