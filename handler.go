package goa

import (
	"net/http"

	"github.com/raphael/goa/design"
)

// Request handler.
type Handler struct {
	// Resource implemented by the controller
	Resource *design.Resource
	// Underlying http response writer
	W http.ResponseWriter
	// Underlying http request
	R *http.Request
}

// Send bad request response, mainly meant to be used by boostrapped code
func (h *Handler) RespondBadRequest(msg string) {
	h.W.WriteHeader(400)
	h.W.Write([]byte(msg))
}
