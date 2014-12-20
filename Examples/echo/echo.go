/*
/ Minimal goa app which implements a single "echo" action
/ Send requests with curl via:
/
/     curl http://localhost:8080/api/echo?value=foo
/
/ Note how goa generates a helpful error response if value param is not provided
*/
package main

import (
	. "github.com/raphael/goa"
	"log"
	"net/http"
	"os"
)

// minimal resource - only one action: "echo"
var resource = Resource{
	RoutePrefix: "/echo",
	Actions: Actions{
		"echo": Action{
			Route: GET("?value={value}"), // Capture param in "value"
			Params: Params{
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
	l := log.New(os.Stdout, "[echo] ", 0)
	addr := "localhost:8080"
	l.Printf("listening on %s", addr)
	l.Printf("(curl http://%s/api/echo?value=foo)", addr)
	l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
}
