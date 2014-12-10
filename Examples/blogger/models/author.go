package model

// Blog post and comment author
type Author struct {
	Id        string `attribute:"id"`
	FirstName string `attribute:"firstName"`
	LastName  string `attribute:"lastName"`
}
