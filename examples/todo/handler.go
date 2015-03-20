package main

import (
	"strconv"

	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/todo/design"
	"github.com/raphael/goa/status"
)

// Task controller implements task resource
type TaskHandler struct {
	*goa.Handler
}

// Task owner data structure
type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// Incoming payload struct for create and update
type TaskDetails struct {
	Owner     User   `json:"owner"`
	Details   string `json:"details"`
	Kind      string `json:"kind"`
	ExpiresAt string `json:"expiresAt"`
}

// Index
func (c *TaskHandler) Index() *goa.Response {
	models := db.LoadAll()
	body, err := design.TaskIndexMediaType.Render(models, "default")
	if err != nil {
		return status.InternalError().WithBody(err)
	}
	return status.Ok().WithBody(body)
}

// Show
func (c *TaskHandler) Show(id int) *goa.Response {
	m := db.Load(id)
	if m == nil {
		return ResourceNotFound(id, "tasks")
	}
	return status.Ok().WithBody(m)
}

// Create
func (c *TaskHandler) Create(p *TaskDetails) *goa.Response {
	id := db.Create(p.Details)
	return status.Created().WithLocation("/tasks/" + strconv.Itoa(int(id)))
}

// Update (upsert semantic)
func (c *TaskHandler) Update(body *TaskDetails, id int) *goa.Response {
	newId := db.Update(id, body.Details)
	if newId != id {
		return status.Created().WithLocation("/tasks/" + strconv.Itoa(int(newId)))
	} else {
		return status.NoContent()
	}
}

// Delete
func (c *TaskHandler) Delete(id int) *goa.Response {
	if db.Delete(id) == 0 {
		return ResourceNotFound(id, "tasks")
	}
	return status.NoContent()
}
