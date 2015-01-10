package main

import (
	"github.com/raphael/goa"
)

// Hello media type
// Identifier: "application/vnd.example.hello"
type HelloMediaType struct {
	goa.MediaType

	// Hello string identifier
	Id uint `goa:"id,MinValue:1,Views:default,tiny"`

	// Hello string content
	Text string `goa:"text,MinLength:1"`
}

// Index response media type
// Identifier: "application/vnd.exampl.hello;type=collection"
type IndexMediaType struct {
  goa.MediaType
  
  // Total number of items
  Count uint `goa:"count"`
  // Items
  Items []TinyHello `goa:"items"`
}

// Tiny view, only list ids
type TinyHello struct {
	// Hello string identifier
	Id uint `goa:"id"`
}

// Hello string, used by create request body
type HelloPayload struct {
	// Value
	Value string `goa:"value,Required:true,MinLength:1"`
}

// Actions  
type HelloActions interface {

	// List all hello strings
	// GET ""
	// 200 ok: body contains JSON array of hello strings
	Index() *IndexMediaType

	// Get hello string with given id
	// GET "/:id"
	// 200: body contains JSON hello string
	// 404: Hello with given id not found
	Show(id uint) *Hello

	// Create new hello string
	// POST ""
	// 201: hello string successfully created
	//      header "Location": href to newly created foo
	Create(body *HelloString)

	// Update existing hello string text
	// PUT "/:id"
	// 204: Successfully updated
	// 404: Hello string with given id not found
	Update(body *HelloString, id uint)

	// Delete hello string
	// DELETE "/:id"
	// 204: Successfully updated
	// 404: Hello string with given id not found
	Delete(id uint)
}
