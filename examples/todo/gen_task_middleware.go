package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa"
	"github.com/raphael/goa/design"
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
	resp := h.Index()
	ok := resp.Status == 400 || resp.Status == 500
	if resp.Status == 200 {
		ok = true
		resp.Header.Set("Content-Type", "application/vnd.acme.task;collection+json")
	}
	if !ok {
		goa.RespondInternalError(w, fmt.Sprintf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	resp.Write(w)
}

func ShowTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := NewTaskHandler(w, r)
	id, err := design.Integer.Load(params.ByName("id"))
	if err != nil {
		goa.RespondBadRequest(w, "invalid param 'id': %s", err)
	}
	view, err := design.String.Load(params.ByName("view"))
	if err != nil {
		goa.RespondBadRequest(w, "invalid param 'view': %s", err)
	}
	resp := h.Show(id.(int), view.(string))
	ok := resp.Status == 400 || resp.Status == 500
	if resp.Status == 200 {
		ok = true
		resp.Header.Set("Content-Type", "application/vnd.acme.task+json")
	}

	if !ok {
		goa.RespondInternalError(w, fmt.Sprintf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	resp.Write(w)
}

func CreateTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := NewTaskHandler(w, r)

	// Load payload
	res := design.Resources["Task"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	raw, err := res.Actions["Create"].Payload.Load("payload", string(body))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	var payload CreatePayload
	err = goa.InitStruct(&payload, raw.(map[string]interface{}))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	resp := h.Create(&payload)
	ok := resp.Status == 400 || resp.Status == 500
	if resp.Status == 201 {
		ok = true
	}
	if !ok {
		goa.RespondInternalError(w, fmt.Sprintf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	resp.Write(w)
}

func UpdateTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := NewTaskHandler(w, r)
	id, err := design.Integer.Load(params.ByName("id"))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
	}
	// Load payload
	res := design.Resources["Task"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	raw, err := res.Actions["Update"].Payload.Load("payload", string(body))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	var payload UpdatePayload
	err = goa.InitStruct(&payload, raw.(map[string]interface{}))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	resp := h.Update(&payload, id.(int))
	ok := resp.Status == 400 || resp.Status == 500
	if resp.Status == 204 {
		ok = true
	}
	if !ok {
		goa.RespondInternalError(w, fmt.Sprintf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	resp.Write(w)
}

func DeleteTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := NewTaskHandler(w, r)
	id, err := design.Integer.Load(params.ByName("id"))
	if err != nil {
		goa.RespondBadRequest(w, "invalid param 'id': %s", err)
	}
	resp := h.Delete(id.(int))
	ok := resp.Status == 400 || resp.Status == 500
	if resp.Status == 204 {
		ok = true
	}
	if !ok {
		goa.RespondInternalError(w, fmt.Sprintf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	resp.Write(w)
}
