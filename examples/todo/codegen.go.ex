package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/old_goa"
)

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
	var payload TaskUpdate
	err := decoder.Decode(&payload)
	if err != nil {
		h.RespondBadRequest(fmt.Sprintf("Failed to load body: %s", err))
	}

	// Call controller Update method
	h.Update(id, payload)

	// Send response
	h.WriteResponse(w)
}
