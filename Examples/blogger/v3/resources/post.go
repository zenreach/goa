package v3

import (
	. "github.com/raphael/goa"
)

var requiredHeaders = Headers{
	"content-type":  "application/json",
	"cache-control": "max-age=0",
}

var postListResponse = Response{
	Status:    200,
	MediaType: postListMediaType,
	Headers:   requiredHeaders,
}
var postResponse = Response{
	Status:    200,
	MediaType: postMediaType,
	Headers:   requiredHeaders,
}

var listNotFound = Response{
	Description: "No post with given author",
	Status:      404,
}

var postResource = Resource{

	RoutePrefix: "/v3/posts",

	Controller: postController,

	MediaType: postMediaType,

	Actions: Actions{

		/* list
		/
		/  GET /v3/posts
		*/
		"list": Action{
			Route: GET(""),

			Description: "Retrieves a list of posts.",

			Params: Params{
				"endDate": Attribute{
					Description: "Latest post date to fetch, a date-time with RFC 3339 formatting.",
					Type:        Time,
				},
				"fetchBodies": Attribute{
					Description:  "Whether the body content of posts is included. This should be set to false when the post bodies are not required, to help minimize traffic.",
					Type:         Boolean,
					DefaultValue: true,
				},
				"fetchImages": Attribute{
					Description:  "Whether image URL metadata for each post is included.",
					Type:         Boolean,
					DefaultValue: true,
				},
				"labels": Attribute{
					Description: "Comma-separated list of labels to search for.",
					Type:        CollectionOf(String),
				},
				"maxResults": Attribute{
					Description: "Maximum number of posts to fetch.",
					Type:        Integer,
					MinValue:    1,
				},
				"orderBy": Attribute{
					Description:   "Sort order applied to results.",
					Type:          String,
					AllowedValues: []string{"published", "updated"},
				},
				"pageToken": Attribute{
					Description: "Continuation token if the request is paged.",
					Type:        String,
				},
				"startDate": Attribute{
					Description: "Earliest post date to fetch, a date-time with RFC 3339 formatting.",
					Type:        Time,
				},
				"status": Attribute{
					Description:   "Filter by status.",
					Type:          String,
					AllowedValues: []string{"draft", "live", "scheduled"},
				},
				"view": Attribute{
					Description:   "Requested view.",
					Type:          String,
					AllowedValues: []string{"ADMIN", "AUTHOR", "READER"},
				},
			},

			Responses: Responses{
				"ok":         postListResponse,
				"badRequest": badRequestResponse,
			},
		},

		/* get
		/
		/  GET /v3/posts/1
		*/
		"get": Action{
			Route: GET("/{postId}"),

			Description: "Retrieves one post by post ID.",

			Params: Params{
				"postId": Attribute{
					Description: "The ID of the post.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
				"maxComments": Attribute{
					Description: "Maximum number of comments to retrieve as part of the the post resource. If this parameter is left unspecified, then no comments will be returned.",
					Type:        Integer,
					MinValue:    1,
				},
				"view": Attribute{
					Description:   "Requested view.",
					Type:          String,
					AllowedValues: []string{"ADMIN", "AUTHOR", "READER"},
				},
			},

			Responses: Responses{
				"ok":         postResponse,
				"badRequest": badRequestResponse,
			},
		},

		/* search
		/
		/  GET /v3/posts/search
		*/
		"search": Action{
			Route: GET("/search"),

			Description: "Searches for a post that matches the given query terms.",

			Params: Params{
				"q": Attribute{
					Description: "Query terms to search for.",
					Type:        String,
					Required:    true,
				},
				"fetchBodies": Attribute{
					Description:  "Whether the body content of posts is included. To minimize traffic, set this parameter to false when the post's body content is not required.",
					Type:         Boolean,
					DefaultValue: true,
				},
				"orderBy": Attribute{
					Description:   "Sort order applied to results.",
					Type:          String,
					AllowedValues: []string{"published", "updated"},
				},
			},

			Responses: Responses{
				"ok":         postListResponse,
				"badRequest": badRequestResponse,
			},
		},

		/* insert
		/
		/  POST /v3/posts
		*/
		"insert": Action{
			Route: POST("/"),

			Description: "Adds a post.",

			Params: Params{
				"isDraft": Attribute{
					Description:  "Whether to create the post as a draft",
					Type:         Boolean,
					DefaultValue: false,
				},
			},

			Payload: postAttributes,

			Responses: Responses{
				"ok":           postResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* delete
		/
		/  DELETE /v3/posts/1
		*/
		"delete": Action{
			Route: DELETE("/{postId}"),

			Description: "Deletes a post by ID.",

			Params: Params{
				"postId": Attribute{
					Description: "The ID of the post.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: Responses{
				"ok":           Http.NoContent(),
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* getByPath
		/
		/  GET /v3/posts/bypath
		*/
		"getByPath": Action{
			Route: GET("/bypath"),

			Description: `Retrieves a post by path. The path of a post is the part of the post URL after the host.
For example, a blog post with the URL http://code.blogger.com/2011/09/blogger-json-api-now-available.html
has a path of /2011/09/blogger-json-api-now-available.html.`,

			Params: Params{
				"path": Attribute{
					Description: "Path of the Post to retrieve.",
					Type:        String,
					Required:    true,
				},
				"maxComments": Attribute{
					Description: "Maximum number of comments to retrieve as part of the the post resource. If this parameter is left unspecified, then no comments will be returned.",
					Type:        Integer,
					MinValue:    1,
				},
				"view": Attribute{
					Description:   "Requested view.",
					Type:          String,
					AllowedValues: []string{"ADMIN", "AUTHOR", "READER"},
				},
			},

			Responses: Responses{
				"ok":           postResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* patch
		/
		/  PATCH /v3/posts/1
		*/
		"patch": Action{
			Route: PATCH("/{postId}"),

			Description: "Updates a post. This method supports patch semantics.",

			Params: Params{
				"postId": Attribute{
					Description: "The ID of the post.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Payload: postAttributes,

			Responses: Responses{
				"ok":           postResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* update
		/
		/  PUT /v3/posts/1
		*/
		"update`": Action{
			Route: PUT("/{postId}"),

			Description: "Updates a post.",

			Params: Params{
				"postId": Attribute{
					Description: "The ID of the post.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Payload: postAttributes,

			Responses: Responses{
				"ok":           postResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* publish
		/
		/  POST /v3/posts/1/publish
		*/
		"publish": Action{
			Route: POST("/{postId}/publish"),

			Description: "Publish a draft post.",

			Params: Params{
				"postId": Attribute{
					Description: "The ID of the post.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
				"publishDate": Attribute{
					Description: "The date and time to schedule the publishing of the Post.",
					Type:        Time,
				},
			},

			Responses: Responses{
				"ok":           postResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* revert
		/
		/  POST /v3/posts/1/revert
		*/
		"revert": Action{
			Route: POST("/{postId}/revert"),

			Description: "Revert a published or scheduled post to draft state, which removes the post from the publicly viewable content.",

			Params: Params{
				"postId": Attribute{
					Description: "The ID of the post.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: Responses{
				"ok":           postResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},
	},
}
