package goa

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// WriteResponse serializes the response body with JSON and writes it to the given response writer.
func WriteResponse(w http.ResponseWriter, r *Response) {
	var b []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		RespondInternalError(w, "API bug, failed to read response body: %s", err)
		return
	}
	if len(body) > 0 {
		var err error
		if b, err = json.Marshal(body); err != nil {
			RespondInternalError(w, "API bug, failed to serialize response body: %s", err)
			return
		}
	}
	if len(r.Header) > 0 {
		header := w.Header()
		for n, v := range r.Header {
			header[n] = v
		}
	}
	w.WriteHeader(r.Status)
	w.Write(b)
}

// Send bad request response, for use by boostrapped code
func RespondBadRequest(w http.ResponseWriter, format string, a ...interface{}) {
	Respondf(w, 400, format, a)
}

// Send internal error response, for use by bootstrapped code
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
