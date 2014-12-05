package main

import (
	. "github.com/raphael/goa"
	"log"
	"net/http"
)

/*
/ Minimal goa app which implements a single "echo" action
/ Send requests with curl via:
/
/     curl http://localhost:8080/api/echo?value=foo -H x-api-version:1.0
/
/ Note how goa generates a helpful error response if value param is not provided
*/

var definition = Resource{
	ApiVersion: "1.0",

	RoutePrefix: "/echo",

	Actions: Actions{

		"echo": Action{
			Route: GET("?value={value}"),
			Params: Attributes{
				"value": Attribute{Type: String, Required: true},
			},
			Responses: Responses{
				"default": Response{Status: 200},
			},
		},
	},
}

type EchoController struct{}

func (c *EchoController) Echo(r Request) {
	r.RespondWithBody("default", r.ParamString("value"))
}

func main() {
	app := NewApplication("/api")
	app.Mount(&definition, &EchoController{})

	log.Fatal(http.ListenAndServe(":8080", app))
}
