package goa

import ()

type Controller struct {
}

func (c *Controller) Respond(code int) *Response {
	return &Response{}
}

type Response struct{}

func (r *Response) WithLocation(l string) *Response {
	return r
}
