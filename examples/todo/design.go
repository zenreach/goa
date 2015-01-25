package main

import (
	"github.com/raphael/goa"
	"time"
)

// Task media type
// A task has a unique id, a kind which can be either 'todo' or 'reminder' and
// details. A task also has a creation timestamp and an expiration timestamp.
// (the idea is that todo tasks get deleted after the expiration timestamp while
// reminders trigger a notification). A task is associated with a user using an
// email address.
// A task media type can be rendered using 2 different views:
//   - The "default" view contains all the field contents and is used when
//     retrieving a specific task (via the "Show" action).
//   - The "tiny" view does not include the details and is used when retrieving
//     a list of tasks (via the "Index" action).
//
//@goa MediaType: "application/vnd.example.todo.task"
type Task struct {
	// Task identifier
	Id uint `goa:"MinValue:1,Views:default,tiny"`

	// User email
	User string `goa:"Pattern:/^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$/"`

	// Task details
	Details string `goa:"MinLength:1"`

	// Task kind
	Kind string `goa:"Enum:todo reminder,goa:"Views:default,tiny"`

	// Todo expiration or reminder alarm timestamp
	ExpiresAt string `format:"time.RFC3339",goa:"Views:default,tiny"`

	// Creation timestamp
	CreatedAt string `format:"time.RFC3339",goa:"Views:default,tiny"`
}

// Collection of tasks media type
// Use "tiny" view to render items
//
//@goa MediaType: "application/vnd.example.task;type=collection"
type TaskCollection struct {
	// Total number of tasks
	Count uint `goa:"Views:default,extended"`

	// Tasks
	Items []*Task `goa:"ViewMappings:default=tiny,extended=default"`
}

// Not found error media type
//
//@goa MediaType: "application/vnd.goa.example.todo.errors.notfound"
type ResourceNotFound struct {
	// Id of resource not found
	Id uint `goa:"MinValue:1"`

	// Type of resource not found
	Resource string `goa:"MinLength:1"`
}

// Invalid "since" error media type
//
//@goa MediaType: "application/vnd.goa.example.todo.errors.invalidsince"
type InvalidSince struct {
	// Original since value
	Since string

	// Validation error details
	Error string
}

// Task details, used to by "create" request body
type TaskDetails struct {
	// Task kind
	Kind string `goa:"Enum:todo reminder,Default:todo"`

	// Todo expiration date or reminder alarm
	ExpiresAt string `format:"time.RFC3339"`

	// Task content
	Details string `goa:"Required,MinLength:1"`
}

// Task resource
// Tasks can be indexed, shown, created, updated and deleted.
// The default media type for task actions that return a 200 OK response is the
// "Task" media type defined above. The "Index" action overrides this default to
// return a task collection media type instead.
// The "Create" action returns a location header containing the href of the
// newly created task.
// Both the "Update" and "Delete" actions return responses with status code 204
// and empty bodies.
//
//@goa Resource
//@goa Version: 1.0
//@goa MediaType: Task
//@goa BasePath: /tasks
type TaskResource interface {

	// List all tasks optionally filtering only the ones created since
	// given date if any.
	//
	//@goa GET "?[since={since}]"
	//@goa 200: TaskCollection
	//@goa 400: InvalidSince
	Index(since string) (*TaskCollection, *InvalidSince)

	// Get task string with given id
	//
	//@goa GET "/{id}"
	//@goa 200: Task
	//@goa 404: ResourceNotFound
	Show(id uint) (*Task, *ResourceNotFound)

	// Create new task string
	// Return path to newly created task in "Location" header
	//
	//@goa POST ""
	//@goa 201:
	//@goa 201 location: /tasks/\d+
	Create(body *TaskDetails)

	// Update existing task string text
	//
	//@goa PUT "/{id}"
	//@goa 204:
	//@goa 404: ResourceNotFound
	Update(body *TaskDetails, id uint) *ResourceNotFound

	// Delete task string
	//
	//@goa DELETE "/{id}"
	//@goa 204:
	//@goa 404: ResourceNotFound
	Delete(id uint) *ResourceNotFound
}
