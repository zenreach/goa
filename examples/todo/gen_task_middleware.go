package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func TaskRouter() http.Handler {
	router := httprouter.New()
	router.GET("/tasks", IndexTask)
	router.GET("/tasks/:id", ShowTask)
	router.POST("/tasks", CreateTask)
	router.PUT("/tasks/:id", UpdateTask)
	router.DELETE("/tasks/:id", DeleteTask)
	return router
}

func IndexTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := NewTaskHandler(w, r)
	r := h.Index()
	ok := r.Status == 400
	if r.Status == 200 {
		ok = true
		r.Header.Set("Content-Type", "application/vnd.acme.task;collection+json")
	}
	if !ok {
		goa.RespondInternalError(fmt.Printf("API bug, code produced unknown status code %d", r.Status))
		return
	}
	r.Write(w)
}

func ShowTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := goa.NewHandler("Task", w, r)
	id, err := goa.Integer.Load(params.ByName("id"))
	if err != nil {
		h.RespondBadRequest("invalid param 'id': %s", err)
	}
	view, err := goa.String.Load(params.ByName("view"))
	if err != nil {
		h.RespondBadRequest("invalid param 'view': %s", err)
	}
	r := h.Show(id, view)
	ok := r.Status == 400
	if r.Status == 200 {
		ok = true
		r.Header.Set("Content-Type", "application/vnd.acme.task+json")
	}

	if !ok {
		goa.RespondInternalError(fmt.Printf("API bug, code produced unknown status code %d", r.Status))
		return
	}
	r.Write(w)
}

func UpdateTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Initialize controller
	h := goa.NewHandler("Task", w, r)

	// Load params
	id, err := goa.Integer.Load(params.ByName("id"))
	if err != nil {
		h.RespondBadRequest(err.Error())
	}

	// Load payload
	decoder := json.NewDecoder(req.Body)
	var payload design.TaskUpdate
	err := decoder.Decode(&payload)
	if err != nil {
		h.RespondBadRequest(fmt.Sprintf("Failed to load body: %s", err))
	}

	// Call controller Update method
	h.Update(id, payload)

	// Send response
	r.Write(w)
}
