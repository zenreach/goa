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
	"./pkg"
	"flag"
	"github.com/raphael/goa"
	"log"
	"net/http"
	"os"
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
		l := log.New(os.Stdout, "[hello] ", 0)
		addr := "localhost:8080"
		l.Printf("listening on %s", addr)
		l.Printf("---------------------------------------------------------------------------------------------------------")
		l.Printf("  index  with `curl http://%s/api/hello`", addr)
		l.Printf("  show   with `curl http://%s/api/hello/1`", addr)
		l.Printf("  update with `curl -X PUT -d '{\"Value\":\"foo\"}' -H Content-Type:application/json http://%s/api/hello/1`", addr)
		l.Printf("  delete with `curl -X DELETE http://%s/api/hello/1`", addr)
		l.Printf("---------------------------------------------------------------------------------------------------------")

		l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
	}
}
