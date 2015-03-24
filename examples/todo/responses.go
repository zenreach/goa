package main

import (
	"github.com/raphael/goa"
)

// Build resource not found response
func ResourceNotFound(id int, name string) *goa.Response {
	body := map[string]interface{}{
		"Id":   id,
		"Name": name,
	}
	return goa.NotFound().WithBody(body)
}
