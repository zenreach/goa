package main

import (
	"github.com/raphael/goa"
)

// Task media type
// Identifier: "application/vnd.example.todo.task"
type Task struct {
	goa.MediaType

	// Task identifier
	Id uint `goa:"MinValue:1,Views:default,tiny"`

	// Task content
	Details string `goa:"MinLength:1"`
}

// Collection of tasks media type
// Use "tiny" view to render items
// Identifier: "application/vnd.example.task;type=collection"
type TaskCollection struct {
	goa.MediaType

	// Total number of tasks
	Count uint

	// Tasks
	Items []Task `goa:"Use:tiny"`
}

// Not found media type
// Identifier: "application/vnd.goa.example.todo.errors.notfound"
type ResourceNotFound struct {
	goa.MediaType

	// Id of resource not found
	Id uint `goa:"MinValue:1"`

	// Type of resource not found
	Resource string `goa:"MinLength:1"`
}

// Task details, used to define create request body
type TaskDetails struct {
	// Task content
	Details string `goa:"Required:true,MinLength:1"`
}

// Actions
type TaskActions interface {

	// List all task strings
	// GET ""
	// 200: TaskCollection
	Index() *TaskCollection

	// Get task string with given id
	// GET "/:id"
	// 200: Task
	// 404: ResourceNotFound
	Show(id uint) (*Task, *ResourceNotFound)

	// Create new task string
	// POST ""
	// 201:
	// Header "location": /task/\d+ // Path to newly created task
	Create(body *TaskDetails)

	// Update existing task string text
	// PUT "/:id"
	// 204
	// 404: ResourceNotFound
	Update(body *TaskDetails, id uint) *ResourceNotFound

	// Delete task string
	// DELETE "/:id"
	// 204
	// 404: ResourceNotFound
	Delete(id uint) *ResourceNotFound
}
