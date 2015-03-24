package goa

import (
	"fmt"
	"net/http"
)

// Send bad request response, intended for generated code
func RespondBadRequest(w http.ResponseWriter, format string, a ...interface{}) {
	Respondf(w, 400, format, a)
}

// Send internal error response, intended for generated code
func RespondInternalError(w http.ResponseWriter, format string, a ...interface{}) {
	Respondf(w, 500, format, a)
}

// Respondf formats the response body using given the format and values and writes the given status
// and resulting string.
func Respondf(w http.ResponseWriter, status int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}
