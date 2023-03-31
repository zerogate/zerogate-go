package zerogate

import "fmt"

const (
	errEmptyCredentials = "API key & secret must not be empty"
)

type Error struct {
	// Response is the error response from the server
	Response ErrorResponse

	// StatusCode is the HTTP status code from the response.
	StatusCode int
}

func (e Error) Error() string {
	if e.Response.ErrorMessage != "" && e.StatusCode > 0 {
		return fmt.Sprintf("%s (%d)", e.Response.ErrorMessage, e.StatusCode)
	}
	return fmt.Sprintf("unknown error (%d)", e.StatusCode)
}
