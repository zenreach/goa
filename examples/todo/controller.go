package main

import (
	"strconv"

	"github.com/raphael/goa"
	"github.com/raphael/goa/status"
)

// Task controller implements task resource
type TaskController struct {
	*goa.Controller
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
func (c *TaskController) Index() *goa.Response {
	models := db.LoadAll()
	body, err := TaskIndexMediaType.Render(models, "default")
	if err != nil {
		return status.InternalError().WithBody(err.Error())
	}
	return status.Ok().WithBody(body)
}

// Show
func (c *TaskController) Show(id int) *goa.Response {
	m := db.Load(id)
	if m == nil {
		return ResourceNotFound(id, "tasks")
	}
	return status.Ok().WithBody(m)
}

// Create
func (c *TaskController) Create(p *TaskDetails) *goa.Response {
	id := db.Create(p.Details)
	return status.Created().WithLocation("/task/" + strconv.Itoa(int(id)))
}

// Update (upsert semantic)
func (c *TaskController) Update(body *TaskDetails, id int) *goa.Response {
	newId := db.Update(id, body.Details)
	if newId != id {
		return status.Created().WithLocation("/tasks/" + strconv.Itoa(int(newId)))
	} else {
		return status.NoContent()
	}
}

// Delete
func (c *TaskController) Delete(id int) *goa.Response {
	if db.Delete(id) == 0 {
		return ResourceNotFound(id, "tasks")
	}
	return status.NoContent()
}
