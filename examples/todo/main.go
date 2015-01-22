// todo example
// Illustrates basic CRUD actions
//
// Run:
//   make run
// Index:
//   curl http://localhost:8080/api/tasks
// Show:
//  curl http://localhost:8080/api/tasks/0
// Create:
//   curl -X POST -d '{"details":"foo"}' http://localhost:8080/api/tasks -H Content-Type:application/json
// Update:
//   curl -X PUT -d '{"details":"foo"}' http://localhost:8080/api/tasks/1 -H Content-Type:application/json
// Delete:
//   curl -X DELETE http://localhost:8080/api/tasks/1
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/raphael/goa"
)

func main() {
	// Setup
	app := goa.New("/api")
	goa_MountAllHandlers(app)

	// Run
	addr := "localhost:8081"
	l := log.New(os.Stdout, "[todo] ", 0)
	l.Printf("Listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, app))
}
