package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/todo/db"
)

// Index
func (t *Task) Index() *TaskCollection {
	return db.LoadAll()
}

// Show
func (t *Task) Show(id uint) *Task, *ResourceNotFound {
	m := db.Load(id)
	if m == nil {
		t.Respond(404)
		return nil, &ResourceNotFound{id, "tasks"}
	}
	return m
}

// Create
func (t *Task) Create(p *TaskDetails) {
	id := db.Create(p.Details)
	t.Respond(201).WithLocation("/task/" + strconv.Itoa(id))
}

// Update (upsert semantic)
func (t *Task) Update(body *TaskDetails, id uint) {
	newId := db.Update(id, body.Details)
	if newId != id {
		t.Respond(201).WithLocation("/tasks/" + strconv.Itoa(newId))
	} else {
		t.Respond(204)
	}
}

// Delete
func (t *Task) Delete(id uint) *ResourceNotFound {
	if db.Delete(id) == 0 {
		t.Respond(404)
		return &ResourceNotFound{id, "tasks"}
	}
	t.Respond(204)
	return nil
}
