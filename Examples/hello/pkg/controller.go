package hello

import (
	"github.com/raphael/goa"
	"strconv"
)

// Hard coded list of hello strings
var greetings = []greeting{
	greeting{1, "Hello world!"},
	greeting{2, "Привет мир!"},
	greeting{3, "Hola mundo!"},
	greeting{4, "你好世界!"},
	greeting{5, "こんにちは世界！"},
}

// Controller
type Hello struct{}

// Index
func (h *Hello) Index(r *goa.Request) {
	r.RespondJson(greetings).
		WithHeader("Content-Type", "application/vnd.example.hello+collection+json")
}

// Show
func (h *Hello) Show(r *goa.Request, id int) {
	found := false
	for _, g := range greetings {
		if g.Id == id {
			found = true
			r.RespondJson(g).
				WithHeader("Content-Type", "application/vnd.example.hello+json")
			break
		}
	}
	if !found {
		r.Respond("").WithStatus(404)
	}
}

// Create
func (h *Hello) Create(r *goa.Request, p *HelloString) {
	new_id := 1
	for ok := false; !ok; new_id += 1 {
		for id, _ := range greetings {
			ok = id != new_id
			if !ok {
				break
			}
		}
	}
	greetings = append(greetings, greeting{new_id, p.Value})
	r.Respond("").WithStatus(201).WithLocation("/hello/" + strconv.Itoa(new_id))
}

// Update
func (h *Hello) Update(r *goa.Request, p *HelloString, id int) {
	found := false
	for idx, g := range greetings {
		if g.Id == id {
			found = true
			greetings[idx] = greeting{id, p.Value}
			break
		}
	}
	if !found {
		greetings = append(greetings, greeting{id, p.Value})
	}
	r.Respond("").WithStatus(204)
}

// Delete
func (h *Hello) Delete(r *goa.Request, id int) {
	found := false
	for i, g := range greetings {
		if g.Id == id {
			greetings[i] = greetings[len(greetings)-1]
			greetings = greetings[0 : len(greetings)-1]
			found = true
			r.Respond("").WithStatus(204)
			break
		}
	}
	if !found {
		r.Respond("").WithStatus(404)
	}
}
