package main

import "strconv"

// Task controller implements task resource
//@goa Controller: TaskResource
type TaskController struct {
	goa.Controller
}

// Index
func (c *TaskController) Index() *TaskCollection {
	models := db.LoadAll()
	tasks := make([]Task, len(models))
	for i, m := range models {
		task, _ := RenderTask(&m)
		tasks[i] = task
	}
	return &TaskCollection{Count: uint(len(tasks)), Items: tasks}
}

// Show
func (c *TaskController) Show(id uint) (*Task, *ResourceNotFound) {
	m := db.Load(id)
	if m == nil {
		c.Respond(404)
		return nil, &ResourceNotFound{Id: id, Resource: "tasks"}
	}
	task, _ := RenderTask(m)
	return task, nil
}

// Create
func (c *TaskController) Create(p *TaskDetails) {
	id := db.Create(p.Details)
	c.Respond(201).WithLocation("/task/" + strconv.Itoa(int(id)))
}

// Update (upsert semantic)
func (c *TaskController) Update(body *TaskDetails, id uint) {
	newId := db.Update(id, body.Details)
	if newId != id {
		c.Respond(201).WithLocation("/tasks/" + strconv.Itoa(int(newId)))
	} else {
		c.Respond(204)
	}
}

// Delete
func (c *TaskController) Delete(id uint) *ResourceNotFound {
	if db.Delete(id) == 0 {
		c.Respond(404)
		return &ResourceNotFound{Id: id, Resource: "tasks"}
	}
	c.Respond(204)
	return nil
}
