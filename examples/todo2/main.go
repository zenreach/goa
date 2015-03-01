//go generate goa
//
// Todo example
// Illustrates basic CRUD actions
//
//   # Build:
//   go generate && go build
//   # Run:
//   ./todo
//   # Create:
//   curl -X POST -d '{"details":"foo"}' http://localhost:8081/api/tasks -H Content-Type:application/json
//   # Index:
//   curl http://localhost:8081/api/tasks
//   # Update:
//   curl -X PUT -d '{"details":"bar"}' http://localhost:8081/api/tasks/1 -H Content-Type:application/json
//   # Show:
//   curl http://localhost:8081/api/tasks/1
//   # Delete:
//   curl -X DELETE http://localhost:8081/api/tasks/1
//
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/raphael/goa"
)

func main() {
	// Define resources
	var task = NewTaskResource()

	// Define controllers
	var taskController = goa.Controller(task)
	
	// Define application
	var app = goa.New("Tasks And Reminder")
	app.Description("Create simple tasks and reminders")
	app.BasePath("/api")
	app.Mount(taskController)

	// Run
	addr := "localhost:8081"
	l := log.New(os.Stdout, "[todo] ", 0)
	l.Printf("Listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, app))
}
