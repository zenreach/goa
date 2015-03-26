package main

import (
	"net/http"
	"strconv"

	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/todo/design"
)

// Task controller struct implements TaskHandler interface.
type taskHandler struct {
}

// Task handler factory, called to handle requests make to task actions
func NewTaskHandler(w http.ResponseWriter, r *http.Request) TaskHandler {
	return &taskHandler{}
}

// Index
func (h *taskHandler) Index() *goa.Response {
	models := db.LoadAll()
	body, err := design.TaskIndexMediaType.Render(models, "default")
	if err != nil {
		return goa.InternalError().WithBody(err)
	}
	return goa.Ok().WithBody(body)
}

// Show
func (h *taskHandler) Show(id int, view string) *goa.Response {
	m := db.Load(id)
	if m == nil {
		return ResourceNotFound(id, "tasks")
	}
	return goa.Ok().WithBody(m)
}

// Create
func (h *taskHandler) Create(p *CreatePayload) *goa.Response {
	id := db.Create(p.Details, p.ExpiresAt)
	return goa.Created().WithLocation("/tasks/" + strconv.Itoa(int(id)))
}

// Update (upsert semantic)
func (h *taskHandler) Update(body *UpdatePayload, id int) *goa.Response {
	newId := db.Update(id, body.Details, body.ExpiresAt)
	if newId != id {
		return goa.Created().WithLocation("/tasks/" + strconv.Itoa(int(newId)))
	} else {
		return goa.NoContent()
	}
}

// Delete
func (h *taskHandler) Delete(id int) *goa.Response {
	if db.Delete(id) == 0 {
		return ResourceNotFound(id, "tasks")
	}
	return goa.NoContent()
}
