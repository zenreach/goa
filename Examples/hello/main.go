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
	"log"
	"net/http"
	"os"
)

func main() {
	// Setup application
	app := goa.New("/api")
	app.Mount("/hello", new(Hello))

	// Run application
	addr := "localhost:8081"
	l := log.New(os.Stdout, "[hello] ", 0)
	l.Printf("Listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
}
