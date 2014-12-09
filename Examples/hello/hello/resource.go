package hello

import . "github.com/raphael/goa"

var HelloResource = Resource{

	MediaType: HelloMediaType,

	ApiVersion: "1.0",

	RoutePrefix: "/hello",

	Actions: Actions{

		"index": Action{
			Description: "List all hello strings",
			Route:       GET(""),
			Responses:   Responses{"ok": Http.Ok()},
		},

		"show": Action{
			Description: "Get hello string with given id",
			Route:       GET("/{id}"),
			Params: Attributes{
				"id": Attribute{Type: Integer, Required: true}, // Inherits other fields from media type attribute
			},
			Responses: Responses{"ok": Http.Ok(), "notFound": Http.NotFound()},
		},

		"update": Action{
			Description: "Replace hello string with given id",
			Route:       PUT("/{id}"),
			Params: Attributes{
				"id": Attribute{Type: Integer, Required: true},
			},
			Payload: Attributes{
				"value": Attribute{
					Type:        String,
					Description: "New value for hello string with given id",
					Required:    true,
				},
			},
			Responses: Responses{"noContent": Http.NoContent(), "notFound": Http.NotFound()},
		},

		"delete": Action{
			Description: "Delete hello string with given id",
			Route:       DELETE("/{id}"),
			Params: Attributes{
				"id": Attribute{Type: Integer, Required: true},
			},
			Responses: Responses{"noContent": Http.NoContent(), "notFound": Http.NotFound()},
		},
	},
}
