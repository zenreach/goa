package main

import (
	"github.com/raphael/goa"
	"time"
)

// Task media type
//@goa MediaType: "application/vnd.example.todo.task"
type Task struct {
	// Task identifier
	Id uint `goa:"MinValue:1,Views:default,tiny"`

	// Task content
	Details string `goa:"MinLength:1"`

	// Creation timestamp
	CreatedAt *time.Time `goa:"Views:default,tiny"`
}

// Collection of tasks media type
// Use "tiny" view to render items
//@goa MediaType: "application/vnd.example.task;type=collection"
type TaskCollection struct {
	// Total number of tasks
	Count uint `goa:"Views:default,extended"`

	// Tasks
	Items []*Task `goa:"ViewMappings:default=tiny,extended=default"`
}

// Not found media type
//@goa MediaType: "application/vnd.goa.example.todo.errors.notfound"
type ResourceNotFound struct {
	// Id of resource not found
	Id uint `goa:"MinValue:1"`

	// Type of resource not found
	Resource string `goa:"MinLength:1"`
}

// Task details, used to define create request body
//@goa Payload
type TaskDetails struct {
	// Task content
	Details string `goa:"Required:true,MinLength:1"`
}

// Task resource
//@goa Resource
//@goa Version: 1.0
//@goa MediaType: Task
type TaskResource interface {
	// List all tasks optionally filtering only the ones created since
	// given date if any.
	//@goa GET "?since={since}"
	//@goa 200: TaskCollection
	Index(since *time.Time) *TaskCollection

	// Get task string with given id
	//@goa GET "/:id"
	//@goa 200: Task
	//@goa 404: ResourceNotFound
	Show(id uint) (*Task, *ResourceNotFound)

	// Create new task string
	// Return path to newly created task in "Location" header
	//@goa POST ""
	//@goa 201:
	//@goa 201 location: /tasks/\d+
	Create(body *TaskDetails)

	// Update existing task string text
	//@goa PUT "/:id"
	//@goa 204:
	//@goa 404: ResourceNotFound
	Update(body *TaskDetails, id uint) *ResourceNotFound

	// Delete task string
	//@goa DELETE "/:id"
	//@goa 204:
	//@goa 404: ResourceNotFound
	Delete(id uint) *ResourceNotFound
}
