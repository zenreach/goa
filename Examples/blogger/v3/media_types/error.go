package v3

import (
	. "github.com/raphael/goa"
)

// Error response data type
var errorType = Composite{
	"domain": Attribute{
		Description: "The error domain.",
		Type:        String,
	},
	"reason": Attribute{
		Description: "The error reason.",
		Type:        String,
	},
	"message": Attribute{
		Description: "The error message.",
		Type:        String,
	},
}

// Error response media type, may contain multiple errors
var errorMediaType = MediaType{

	Identifier: "vnd.example.blogger.error",

	Model: Model{
		Blueprint: model.Error,

		Attributes: Attributes{
			"errors": Attribute{
				Description: "The list of errors.",
				Type:        CollectionOf(errorType),
				Required:    true,
				MinLength:   1,
			},
			"code": Attribute{
				Description: "The error HTTP code.",
				Type:        Integer,
				Required:    true,
				MinValue:    400,
				MaxValue:    599,
			},
			"message": Attribute{
				Description: "The error summary message",
				Type:        String,
				Required:    true,
			},
		},
	},
}
