package v3

import (
	"github.com/raphael/goa"
)

var commentListResponse = goa.Response{
	Status:    200,
	MediaType: commentListMediaType,
	Headers: goa.Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}
var commentResponse = goa.Response{
	Status:    200,
	MediaType: commentMediaType,
	Headers: goa.Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}

var badRequestResponse = goa.Response{
	Status:    400,
	MediaType: errorMediaType,
	Headers: goa.Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}

var unauthorizedResponse = goa.Response{
	Status:    401,
	MediaType: errorMediaType,
	Headers: goa.Headers{
		"content-type":  "~application/json",
		"cache-control": "~max-age=0",
	},
}

var commentSpec = goa.ControllerSpec{

	ApiVersion: "3.0",

	RoutePrefix: "/v3/posts/{postId}/comments",

	Params: goa.Attributes{
		"postId": goa.Attribute{
			Description: "The ID of the post to fetch comments from.",
			Type:        goa.String,
			Required:    true,
			Regexp:      "[0-9]+",
		},
	},

	Controller: commentController,

	MediaType: commentMediaType,

	Actions: goa.Actions{

		/* list
		/
		/  GET /v3/posts/{postId}/comments
		*/
		"list": goa.GET{
			Path: "",

			Description: "Retrieves the list of comments for a post.",

			Params: goa.Attributes{
				"endDate": goa.Attribute{
					Description: "Latest comment date to fetch, a date-time with RFC 3339 formatting.",
					Type:        goa.Datetime,
				},
				"fetchBodies": goa.Attribute{
					Description:  "Whether the body content of comments is included. This should be set to false when the comment bodies are not required, to help minimize traffic.",
					Type:         goa.Boolean,
					DefaultValue: true,
				},
				"maxResults": goa.Attribute{
					Description: "Maximum number of comments to fetch.",
					Type:        goa.Integer,
					MinValue:    1,
				},
				"pageToken": goa.Attribute{
					Description: "Continuation token if the request is paged.",
					Type:        goa.String,
				},
				"startDate": goa.Attribute{
					Description: "Earliest comment date to fetch, a date-time with RFC 3339 formatting.",
					Type:        goa.Datetime,
				},
				"status": goa.Attribute{
					Description: "Filter by status.",
					Type:        goa.String,
					AllowedValues: goa.Values{
						"emptied": "Comments that have had their content removed",
						"live":    "Comments that are publicly visible",
						"pending": "Comments that are awaiting administrator approval",
						"spam":    "Comments marked as spam by the administrator",
					},
				},
				"view": goa.Attribute{
					Description: "Requested view.",
					Type:        goa.String,
					AllowedValues: goa.Values{
						"ADMIN":  "Admin level detail",
						"AUTHOR": "Author level detail",
						"READER": "Reader level detail",
					},
				},
			},

			Responses: goa.Responses{
				"ok":           commentListResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* get
		/
		/  GET /v3/posts/{postId}/comments/{commentId}
		*/
		"get": goa.GET{
			Path: "/{commentId}",

			Description: "Retrieves one comment by comment ID.",

			Params: goa.Attributes{
				"commentId": goa.Attribute{
					Description: "The ID of the comment.",
					Type:        goa.String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: goa.Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* approve
		/
		/  POST /v3/posts/{postId}/comments/{commentId}/approve
		*/
		"approve": goa.POST{
			Path: "/{commentId}/approve",

			Description: "Marks a comment as not spam.",

			Params: goa.Attributes{
				"commentId": goa.Attribute{
					Description: "The ID of the comment to mark as not spam.",
					Type:        goa.String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: goa.Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* delete
		/
		/  DELETE /v3/posts/{postId}/comments/{commentId}
		*/
		"delete": goa.DELETE{
			Path: "/{commentId}",

			Description: "Deletes a comment by ID.",

			Params: goa.Attributes{
				"commentId": goa.Attribute{
					Description: "The ID of the comment.",
					Type:        goa.String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: goa.Responses{
				"ok":           goa.Response{status: 204},
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* markAsSpam
		/
		/  POST /v3/posts/{postId}/comments/{commentId}/spam
		*/
		"markAsSpam": goa.POST{
			Path: "/{commentId}/spam",

			Description: "Marks a comment as spam.",

			Params: goa.Attributes{
				"commentId": goa.Attribute{
					Description: "The ID of the comment to mark as spam.",
					Type:        goa.String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Responses: goa.Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},

		/* removeContent
		/
		/  POST /v3/posts/{postId}/comments/{commentId}/removecontent
		*/
		"removeContent": goa.POST{
			Path: "/{commentId}/removecontent",

			Description: "Removes the content of a comment.",

			Params: goa.Attributes{
				"commentId": goa.Attribute{
					Description: "The ID of the comment.",
					Type:        goa.String,
					Required:    true,
					Regexp:      "[0-9]+",
				},
			},

			Payload: commentMediaType.GetAttributes(),

			Responses: goa.Responses{
				"ok":           commentResponse,
				"badRequest":   badRequestResponse,
				"unauthorized": unauthorizedResponse,
			},
		},
	},
}
