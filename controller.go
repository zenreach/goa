package goa

import ()

type Controller struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Controller) Respond(code int) *Response {
	return &Response{}
}

type Response struct{}

func (r *Response) WithLocation(l string) *Response {
	return r
}
