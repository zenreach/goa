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
//@goa MediaType: application/vnd.example.todo.task
type Task struct {
	// Task identifier
	Id uint `goa:"minValue:1,views:default tiny"`

	// User email
	User string `goa:"pattern:/^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$/"`

	// Task details
	Details string `goa:"minLength:1"`

	// Task kind
	Kind string `goa:"enum:todo reminder,views:default tiny"`

	// Todo expiration or reminder alarm timestamp
	ExpiresAt string `goa:"views:default tiny,format:time.RFC3339"`

	// Creation timestamp
	CreatedAt string `goa:"views:default tiny,format:time.RFC3339"`
}

// Collection of tasks media type
// Use "tiny" view to render items
//
//@goa MediaType: application/vnd.example.task;type=collection
type TaskCollection struct {
	// Total number of tasks
	Count uint `goa:"views:default extended"`

	// Tasks
	Items []*Task `goa:"viewMappings:default=tiny extended=default"`
}

// Not found error media type
//
//@goa MediaType: application/vnd.goa.example.todo.errors.notfound
type ResourceNotFound struct {
	// Id of resource not found
	Id uint `goa:"minValue:1"`

	// Type of resource not found
	Resource string `goa:"minLength:1"`
}

// Invalid "since" error media type
//
//@goa MediaType: application/vnd.goa.example.todo.errors.invalidsince
type InvalidSince struct {
	// Original since value
	Since string

	// Validation error details
	Error string
}

// Task details, used to by "create" request body
type TaskDetails struct {
	// Task kind
	Kind string `goa:"enum:todo reminder,default:todo"`

	// Todo expiration date or reminder alarm
	ExpiresAt string `format:"time.RFC3339"`

	// Task content
	Details string `goa:"required,minLength:1"`
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
//@goa Name: tasks
//@goa Version: 1.0
//@goa MediaType: application/vnd.example.task
//@goa BasePath: /tasks
type TaskResource interface {

	// List all tasks optionally filtering only the ones created since
	// given date if any.
	//
	//@goa GET "?[since={since}]"
	//@goa Action: index
	//@goa Views: tiny
	//@goa 200: application/vnd.example.todo.task;type=collection
	//@goa 400: application/vnd.goa.example.todo.errors.invalidsince
	Index(since string) (*TaskCollection, *InvalidSince)

	// Get task string with given id
	//
	//@goa GET "/{id}"
	//@goa Action: show
	//@goa Views: tiny, default
	//@goa 200: application/vnd.example.todo.task
	//@goa 404: application/vnd.goa.example.todo.errors.notfound
	Show(id uint) (*Task, *ResourceNotFound)

	// Create new task string
	// Return path to newly created task in "Location" header
	//
	//@goa POST ""
	//@goa Action: create
	//@goa 201:
	//@goa 201 location: /tasks/\d+
	Create(body *TaskDetails)

	// Update existing task string text
	//
	//@goa PUT "/{id}"
	//@goa Action: update
	//@goa 204:
	//@goa 404: application/vnd.goa.example.todo.errors.notfound
	Update(body *TaskDetails, id uint) *ResourceNotFound

	// Delete task string
	//
	//@goa DELETE "/{id}"
	//@goa Action: delete
	//@goa 204:
	//@goa 404: application/vnd.goa.example.todo.errors.notfound
	Delete(id uint) *ResourceNotFound
}
