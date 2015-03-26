package main

import (
	"time"

	"github.com/raphael/goa"
)

// Task handler interface
type TaskHandler interface {
	Index() *goa.Response
	Show(id int, view string) *goa.Response
	Create(payload *CreatePayload) *goa.Response
	Update(payload *UpdatePayload, id int) *goa.Response
	Delete(id int) *goa.Response
}

// Task owner data structure
type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// Incoming payload struct for create
type CreatePayload struct {
	Owner     *User     `json:"owner"`
	Details   string    `json:"details"`
	Kind      string    `json:"kind"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Incoming payload struct for update
type UpdatePayload struct {
	Details   string    `json:"details"`
	ExpiresAt time.Time `json:"expiresAt"`
}
