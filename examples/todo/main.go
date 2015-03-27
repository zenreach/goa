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
//   curl -X POST -d '{"details":"foo"}' http://localhost:8081/tasks -H Content-Type:application/json
//   # Index:
//   curl http://localhost:8081/api/tasks
//   # Update:
//   curl -X PUT -d '{"details":"bar"}' http://localhost:8081/tasks/1 -H Content-Type:application/json
//   # Show:
//   curl http://localhost:8081/tasks/1
//   # Delete:
//   curl -X DELETE http://localhost:8081/tasks/1
//
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/todo/design"
)

func main() {
	// Initialize resources
	design.Init()

	// Define application
	app := goa.New("Tasks And Reminder", "Create simple tasks")
	loggingTaskHandler := handlers.LoggingHandler(os.Stdout, TaskRouter())
	app.Mount("/tasks", loggingTaskHandler)

	// Run
	addr := "localhost:8081"
	log.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, app))
}
