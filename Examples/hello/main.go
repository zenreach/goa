// hello world example
// Shows basic usage of goa
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
