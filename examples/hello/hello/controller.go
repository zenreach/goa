package hello

import "github.com/raphael/goa"

// Hard coded list of hello strings
var greetings = []greeting{
	greeting{0, "Hello world!"},
	greeting{1, "Привет мир!"},
	greeting{2, "Hola mundo!"},
	greeting{3, "你好世界!"},
	greeting{4, "こんにちは世界！"},
}

// Controller
type Hello struct{}

// Index
func (h *Hello) Index(r goa.Request) {
	r.RespondWithBody("ok", greetings)
}

// Show
func (h *Hello) Show(r goa.Request) {
	found := false
	for _, g := range greetings {
		if g.Id == int(r.ParamInt("id")) {
			found = true
			r.RespondWithBody("ok", g)
			break
		}
	}
	if !found {
		r.RespondEmpty("not_found")
	}
}

// Update
func (h *Hello) Update(r goa.Request) {
	found := false
	id := int(r.ParamInt("id"))
	for idx, g := range greetings {
		if g.Id == id {
			found = true
			greetings[idx] = greeting{id, r.PayloadString("value")}
			break
		}
	}
	if !found {
		greetings = append(greetings, greeting{id, r.PayloadString("value")})
	}
	r.RespondEmpty("no_content")
}

// Delete
func (h *Hello) Delete(r goa.Request) {
	found := false
	id := int(r.ParamInt("id"))
	for i, g := range greetings {
		if g.Id == id {
			greetings[i] = greetings[len(greetings)-1]
			greetings = greetings[0 : len(greetings)-1]
			found = true
			r.RespondEmpty("no_content")
			break
		}
	}
	if !found {
		r.RespondEmpty("not_found")
	}
}
