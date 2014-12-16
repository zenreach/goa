// hello world example
// Shows basic usage of goa
//
// Run:
//   make run
// Index:
//   curl http://localhost:8080/api/hello -H x-api-version:1.0
// Show:
//  curl http://localhost:8080/api/hello/0 -H x-api-version:1.0
// Update:
//   curl -X PUT -d '{"Value":"foo"}' http://localhost:8080/api/hello/0 -H x-api-version:1.0  -H Content-Type:application/json
// Delete:
//   curl -X DELETE http://localhost:8080/api/hello/0
package main

import (
	"./hello"
	"flag"
	"github.com/raphael/goa"
	"net/http"
)

func main() {
	// Setup --routes flag
	printRoutes := flag.Bool("routes", false, "Print routes")
	flag.Parse()

	// Setup application
	app := goa.NewApplication("/api")
	app.Mount(&hello.HelloResource, &hello.Hello{})

	// Print routes or run app
	if *printRoutes {
		app.PrintRoutes()
	} else {
		http.ListenAndServe(":8080", app)
	}
}
