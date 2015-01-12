package goa

import (
	"net/http"
)

type Controller struct {
	Respond(code int) *Response
}

func (c *Controller) Respond(code int) *Response {
	return &Response{}
}

type Response struct{}

func (r *Response) WithLocation(l string) *Response {
	return r
}