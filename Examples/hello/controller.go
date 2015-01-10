package main

import (
	"github.com/raphael/goa"
)

// Hard coded pre-existing list of hello strings
var hellos = []Hello{
	Hello{1, "Hello world!"},
	Hello{2, "Привет мир!"},
	Hello{3, "Hola mundo!"},
	Hello{4, "你好世界!"},
	Hello{5, "こんにちは世界！"},
}

// Index
func (h *Hello) Index() []Hello {
	return hellos
}

// Render "tiny" view for index
func (h *Hello) FormatIndex(hellos []Hello) string {
	tinyHellos := make([]TinyHello, len(hellos))
	for i, h := range hellos {
		tinyHellos[i] = goa.View(h, "tiny")
	}
	m := IndexMediaType{len(h), tinyHellos}
	return m.Render()
}

// Show
func (h *Hello) Show(id uint) *Hello {
	found := false
	for _, h := range hellos {
		if h.Id == id {
			return h
		}
	}
	h.Respond(404)
}

// Create
func (h *Hello) Create(p *HelloString) {
	newId := 1
	for ok := false; !ok; newId += 1 {
		for id, _ := range hellos {
			ok = id != newId
			if !ok {
				break
			}
		}
	}
	Hellos = append(Hellos, Hello{newId, p.Value})
	h.Respond(201).WithLocation("/hello/" + strconv.Itoa(newId))
}

// Update
func (h *Hello) Update(body *HelloString, id uint) {
  	found := false
	for idx, h := range hellos {
		if h.Id == id {
			found = true
			hellos[idx] = hellos{id, p.Value}
			break
		}
	}
	if !found {
		hellos = append(hellos, hello{id, p.Value})
	}
	h.Respond(204)
}

// Delete
func (h *Hello) Delete(id uint) {
	found := false
	for i, h := range hellos {
		if g.Id == id {
			hellos[i] = hellos[len(hellos)-1]
			hellos = hellos[0 : len(hellos)-1]
			found = true
			h.Respond(204)
			break
		}
	}
	if !found {
		h.Respond(404)
	}
}
}
