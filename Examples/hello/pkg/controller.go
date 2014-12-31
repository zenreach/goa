package hello

import "github.com/raphael/goa"

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
		WithHeader("Content-Type", "application/vnd.example.hello+json")
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

// Update
func (h *Hello) Update(r *goa.Request, payload *TValue, id int) {
	found := false
	for idx, g := range greetings {
		if g.Id == id {
			found = true
			greetings[idx] = greeting{id, payload.Value}
			break
		}
	}
	if !found {
		greetings = append(greetings, greeting{id, payload.Value})
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
