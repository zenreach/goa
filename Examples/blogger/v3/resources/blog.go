package v3

import (
	. "github.com/raphael/goa"
)

var requiredHeaders = Headers{
	"content-type":  "application/json",
	"cache-control": "max-age=0",
}

var blogResource = Resource{

	RoutePrefix: "/v3/blogs",

	MediaType: blogMediaType,

	Actions: Actions{

		/* list
		/
		/  GET /v3/blogs/{blogId}
		*/
		"list": Action{
			Route: GET("/{blogId}"),

			Description: "Retrieves a blog by its ID.",

			Params: Params{
				"blogId": Attribute{
					Description: "The ID of the blog to get.",
					Type:        String,
					Required:    true,
				},
				"maxPosts": Attribute{
					Description: "Maximum number of posts to retrieve along with the blog." +
						" When this parameter is not specified, no posts will be" +
						" returned as part of the blog resource.",
					Type:     Integer,
					MinValue: 1,
				},
			},

			Responses: Responses{
				"ok":       http.Ok().WithResourceCollection(),
				"notFound": http.NotFound(),
			},
		},

		/* getByUrl
		/
		/  GET /v3/blogs/byurl
		*/
		"get": Action{
			Route: GET("/byurl"),

			Description: "Retrieves a blog by URL.",

			Params: Params{
				"url": Attribute{
					Description: "The URL of the blog to retrieve.",
					Type:        String,
					Required:    true,
				},
			},

			Responses: Responses{
				"ok": blogResponse,
			},
		},
	},
}
