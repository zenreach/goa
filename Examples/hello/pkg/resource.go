package hello

import . "github.com/raphael/goa"

// Update payload data type
type HelloString struct {
	Value string `attribute:"value"`
}

var HelloResource = Resource{

	Name: "Hello",

	RoutePrefix: "/hello",

	MediaType: HelloMediaType,

	Actions: Actions{

		"index": Action{
			Description: "List all hello strings",
			Route:       GET(""),
			Responses:   Responses{"ok": Http.Ok().WithResourceCollection()},
		},

		"show": Action{
			Description: "Get hello string with given id",
			Route:       GET("/{id}"),
			Params: Params{
				"id": Attribute{Type: Integer, Required: true}, // Inherits other fields from media type attribute
			},
			Responses: Responses{"ok": Http.Ok().WithResource(),
				"notFound": Http.NotFound()},
		},

		"create": Action{
			Description: "Create new hello string",
			Route:		 POST(""),
			Payload: Payload{
				Blueprint: HelloString{},
				Attributes: Attributes{
					"value": Attribute{Type: String, Required: true},
				},
			},
			Responses: Responses{"created": Http.Created().
				WithLocation("//hello/[1-9]+/")},
		},

		"update": Action{
			Description: "Replace hello string with given id",
			Route:       PUT("/{id}"),
			Params: Params{
				"id": Attribute{Type: Integer, Required: true},
			},
			Payload: Payload{
				Blueprint: HelloString{},
				Attributes: Attributes{
					"value": Attribute{
						Type:        String,
						Description: "New value for hello string with given id",
						Required:    true,
					},
				},
			},
			Responses: Responses{"noContent": Http.NoContent(), "notFound": Http.NotFound()},
		},

		"delete": Action{
			Description: "Delete hello string with given id",
			Route:       DELETE("/{id}"),
			Params: Params{
				"id": Attribute{Type: Integer, Required: true},
			},
			Responses: Responses{"noContent": Http.NoContent(), "notFound": Http.NotFound()},
		},
	},
}
