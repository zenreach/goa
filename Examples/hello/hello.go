// hello world example
// Shows basic usage of goa
//
// Run:
//   make run
// Index:
//   curl http://localhost:8080/api/hello
// Show:
//  curl http://localhost:8080/api/hello/0
// Update:
//   curl -X PUT -d '{"Value":"foo"}' http://localhost:8080/api/hello/1 -H Content-Type:application/json
// Delete:
//   curl -X DELETE http://localhost:8080/api/hello/1
package main

import (
	"github.com/raphael/goa"
	. "github.com/raphael/goa/Examples/hello/pkg"
	"log"
	"net/http"
	"os"
)

func main() {
	// Setup application
	app := goa.New("/api")
	app.Mount(&Hello{}, &HelloResource)

	// Print docs, routes and run app
	addr := "localhost:8081"
	docs := goa.GenerateSwagger(app, info(), addr)
	l := log.New(os.Stdout, "[hello] ", 0)
	l.Printf("Listening on %s", addr)
	l.Println("Docs:")
	l.Println(docs)
	l.Println("Routes:")
	app.Routes().Log(l)
	l.Printf("\n")
	l.Printf("---------------------------------------------------------------------------------")
	l.Printf("  index   `curl http://%s/api/hello`", addr)
	l.Printf("  show    `curl http://%s/api/hello/1`", addr)
	l.Printf("  create: `curl -X POST -d '{\"value\":\"foo\"}'\\\n"+
		"                   -H 'Content-Type:application/json' http://%s/api/hello`", addr)
	l.Printf("  update: `curl -X PUT -d '{\"value\":\"foo\"}'\\\n"+
		"                   -H 'Content-Type:application/json' http://%s/api/hello/1`", addr)
	l.Printf("  delete: `curl -X DELETE http://%s/api/hello/1`", addr)
	l.Printf("---------------------------------------------------------------------------------")

	l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
}

// Information used to generate Swagger docs
func info() *goa.SwaggerInfo {
	return &goa.SwaggerInfo{
		Title: "goa *hello* example",
		Description: "Simple example that demonstrates basic CRUD" +
			" operations on a REST resource",
		Contact: &goa.SwaggerContact{
			Name:  "Raphael Simon",
			Url:   "https://github.com/raphael/goa",
			Email: "simon.raphael@gmail.com",
		},
		License: &goa.SwaggerLicense{Name: "MIT"},
		Version: "v1.0",
	}
}
