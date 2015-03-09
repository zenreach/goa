package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/status"
)

// Build resource not found response
func ResourceNotFound(id int, name string) *goa.Response {
	body := map[string]interface{}{
		"Id":   id,
		"Name": name,
	}
	return status.NotFound().WithBody(body)
}
