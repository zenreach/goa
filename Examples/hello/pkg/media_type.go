package hello

import . "github.com/raphael/goa"

// Go type used to build media type instances
type greeting struct {
	Id   int    `json:"id" attribute:"id"`
	Text string `json:"text" attribute:"text"`
}

// Media type definition
var HelloMediaType = MediaType{

	Identifier: "application/vnd.example.hello+json",

	Model: Model{
		Blueprint: greeting{},
		Attributes: Attributes{
			"id": Attribute{
				Type:        Integer,
				Description: "Hello string identifier",
				MinValue:    0,
			},
			"text": Attribute{
				Type:        String,
				Description: "Hello string content",
				MinLength:   1,
			},
		},
	},

	Views: Views{
		"default": View{
			Description: "Default view",
			Attributes: Attributes{
				"id":   Attribute{},
				"text": Attribute{},
			},
		},
	},
}
