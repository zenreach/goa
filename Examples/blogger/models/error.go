package model

type Error struct {
	Domain  string `attribute:"domain"`
	Reason  string `attribute:"reason"`
	Message string `attribute:"message"`
}

type Errors struct {
	Errors  []*Error `attribute:"errors"`
	Code    int      `attribute:"code"`
	Message string   `attribute:"message"`
}
