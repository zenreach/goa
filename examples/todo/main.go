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
	"github.com/raphael/goa"
	"log"
	"net/http"
	"os"
)

func main() {
	// Setup
	app := goa.New("/api")
	app.Mount("/tasks", new(TaskController))

	// Run
	addr := "localhost:8081"
	l := log.New(os.Stdout, "[todo] ", 0)
	l.Printf("Listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
}
