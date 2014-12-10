package v3

import (
	. "github.com/raphael/goa"
)

var commentListResponse = Response{
	Status:    200,
	MediaType: commentListMediaType,
	Headers: Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}
var commentResponse = Response{
	Status:    200,
	MediaType: commentMediaType,
	Headers: Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}

var badRequestResponse = Response{
	Status:    400,
	MediaType: errorMediaType,
	Headers: Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}

var unauthorizedResponse = Response{
	Status:    401,
	MediaType: errorMediaType,
	Headers: Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}

var commentSpec = ControllerSpec{

	ApiVersion: "3.0",

	RoutePrefix: "/v3/posts/{postId}/comments",

	Params: Params{
		"postId": Attribute{
			Description: "The ID of the post to fetch comments from.",
			Type:        String,
			Required:    true,
			Regexp:      "[0-9]+",
		},
	},

	Controller: commentController,

	MediaType: commentMediaType,

	Actions: Actions{

		/* list
		/
		/  GET /v3/posts/{postId}/comments
		*/
		"list": Action{
			Route: GET(""),

			Description: "Retrieves the list of comments for a post.",

			Params: Params{
				"endDate": Attribute{
					Description: "Latest comment date to fetch, a date-time with RFC 3339 formatting.",
					Type:        Datetime,
				},
				"fetchBodies": Attribute{
					Description:  "Whether the body content of comments is included. This should be set to false when the comment bodies are not required, to help minimize traffic.",
					Type:         Boolean,
					DefaultValue: true,
				},
				"maxResults": Attribute{
					Description: "Maximum number of comments to fetch.",
					Type:        Integer,
					MinValue:    1,
				},
				"pageToken": Attribute{
					Description: "Continuation token if the request is paged.",
					Type:        String,
				},
				"startDate": Attribute{
					Description: "Earliest comment date to fetch, a date-time with RFC 3339 formatting.",
					Type:        Datetime,
				},
				"status": Attribute{
					Description: "Filter by status.",
					Type:        String,
					AllowedValues: Values{
						"emptied": "Comments that have had their content removed",
						"live":    "Comments that are publicly visible",
						"pending": "Comments that are awaiting administrator approval",
						"spam":    "Comments marked as spam by the administrator",
					},
				},
				"view": Attribute{
					Description: "Requested view.",
					Type:        String,
					AllowedValues: Values{
						"ADMIN":  "Admin level detail",
						"AUTHOR": "Author level detail",
						"READER": "Reader level detail",
					},
				},
			},

			Responses: Responses{
				"ok":           commentListResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* get
		/
		/  GET /v3/posts/{postId}/comments/{commentId}
		*/
		"get": Action{
			Route: GET("/{commentId}"),

			Description: "Retrieves one comment by comment ID.",

			Params: Params{
				"commentId": Attribute{
					Description: "The ID of the comment.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* approve
		/
		/  POST /v3/posts/{postId}/comments/{commentId}/approve
		*/
		"approve": Action{
			Route: POST("/{commentId}/approve"),

			Description: "Marks a comment as not spam.",

			Params: Params{
				"commentId": Attribute{
					Description: "The ID of the comment to mark as not spam.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* delete
		/
		/  DELETE /v3/posts/{postId}/comments/{commentId}
		*/
		"delete": Action{
			Route: DELETE("/{commentId}"),

			Description: "Deletes a comment by ID.",

			Params: Params{
				"commentId": Attribute{
					Description: "The ID of the comment.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: Responses{
				"ok":           Response{status: 204},
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* markAsSpam
		/
		/  POST /v3/posts/{postId}/comments/{commentId}/spam
		*/
		"markAsSpam": Action{
			Route: POST("/{commentId}/spam"),

			Description: "Marks a comment as spam.",

			Params: Params{
				"commentId": Attribute{
					Description: "The ID of the comment to mark as spam.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* removeContent
		/
		/  POST /v3/posts/{postId}/comments/{commentId}/removecontent
		*/
		"removeContent": Action{
			Route: POST("/{commentId}/removecontent"),

			Description: "Removes the content of a comment.",

			Params: Params{
				"commentId": Attribute{
					Description: "The ID of the comment.",
					Type:        String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Payload: commentMediaType.GetAttributes(),

			Responses: Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},
	},
}
