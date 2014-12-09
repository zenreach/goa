/*
/ Minimal goa app which implements a single "echo" action
/ Send requests with curl via:
/
/     curl http://localhost:8080/api/echo?value=foo -H x-api-version:1.0
/
/ Note how goa generates a helpful error response if value param is not provided
*/
package main

import (
	. "github.com/raphael/goa"
	"net/http"
)

// minimal resource - only one action: "echo"
var resource = Resource{
	ApiVersion:  "1.0",
	RoutePrefix: "/echo",
	Actions: Actions{
		"echo": Action{
			Route: GET("?value={value}"), // Capture param in "value"
			Params: Attributes{
				"value": Attribute{Type: String, Required: true},
			},
			Responses: Responses{
				"ok": Response{Status: 200}, // Only one response code possible
			},
		},
	},
}

// Controller struct, minimal doesn't need state so empty
type EchoController struct{}

// Action implementation
func (c *EchoController) Echo(r Request) {
	r.RespondWithBody("ok", r.ParamString("value")) // Send default response, use "value" param as response body
}

// Entry point
func main() {
	app := NewApplication("/api")           // Create application
	app.Mount(&resource, &EchoController{}) // Mount resource and corresponding controller
	http.ListenAndServe(":8080", app)       // Application implements standard http.Handlefunc
}
