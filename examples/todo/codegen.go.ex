import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/old_goa"
)

func UpdateTask(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Initialize controller
	c := &TaskController{w: w, r: r}

	// Load params
	id, err := goa.Integer.Load(params.ByName("id"))
	if err != nil {
		c.RespondBadRequest(err.Error())
	}

	// Load payload
	decoder := json.NewDecoder(req.Body)
	var payload TaskUpdate
	err := decoder.Decode(&payload)
	if err != nil {
		c.RespondBadRequest(fmt.Sprintf("Failed to load body: %s", err.Error()))
	}

	// Call controller Update method
	c.Update(id, payload)

	// Send response
	c.WriteResponse(w)
}
