package v3

import (
	. "github.com/raphael/goa"
)

// List post by user response media type
var postListMediaType = MediaType{

	Identifier: "vnd.example.blogger.postList",

	Model: Model{
		Blueprint: model.PostList{},

		Attributes: Attributes{
			"kind": Attribute{
				Description: "The kind of this entity. Always blogger#postList",
				Type:        String,
			},
			"nextPageToken": Attribute{
				Description: "Pagination token to fetch the next page, if one exists.",
				Type:        String,
			},
			"items": Attribute{
				Description: "The list of posts for this blog.",
				Type:        CollectionOf(postMediaType),
			},
		},
	},
}

// Blog post media type
var postMediaType = MediaType{

	Identifier: "vnd.example.blogger.post",

	Model: Model{
		Blueprint: model.Post{},

		Attributes: Attributes{
			"kind": Attribute{
				Description: "The kind of this resource. Always goa#blogger#post",
				Type:        String,
				Regexp:      "^goa#blogger#post$",
			},
			"id": Attribute{
				Description: "The ID for this resource.",
				Type:        String,
				Regexp:      "[0-9]+",
			},
			"blog": Attribute{
				Description: "",
				Type:        blogMediaType,
				View:        "tiny",
			},
			"published": Attribute{
				Description: "date-time when this post was published.",
				Type:        Datetime,
			},
			"updated": Attribute{
				Description: "date-time when this post was last updated.",
				Type:        Datetime,
			},
			"url": Attribute{
				Description: "The URL where this post is displayed.",
				Type:        String,
			},
			"selfLink": Attribute{
				Description: "The Blogger API URL to fetch this resource from.",
				Type:        String,
			},
			"title": Attribute{
				Description: "The title of the post.",
				Type:        String,
				MinLength:   10,
				MaxLength:   1024,
			},
			"titleLink": Attribute{
				Description: "The title link URL, similar to atom's related link.",
				Type:        String,
			},
			"content": Attribute{
				Description: "The content of the post. Can contain HTML markup.",
				Type:        String,
				MinLength:   255,
			},
			"author": Attribute{
				Description: "Blog author",
				Type:        authorType,
			},
			"replies": Attribute{
				Description: "The container for this post's comments.",
				Type:        repliesType,
			},
			"labels": Attribute{
				Description: "The list of labels this post was tagged with.",
				Type:        CollectionOf(String),
			},
			"customMetaData": Attribute{
				Description: "The JSON metadata for the post.",
				Type:        Json,
			},
			"location": Attribute{
				Description: "The location, if this post is geotagged.",
				Type:        locationType,
			},
			"images": Attribute{
				Description: "Display image for the Post.",
				Type:        CollectionOf(imageType),
			},
			"status": Attribute{
				Description:   "Status of the post. Only set for admin-level requests",
				Type:          String,
				AllowedValues: []string{"draft", "live", "scheduled"},
			},
		},

		Views: Views{
			"reader": View{
				Description: "Reader level detail",
				Attributes: Attributes{
					"kind":           Attribute{},
					"id":             Attribute{},
					"published":      Attribute{},
					"url":            Attribute{},
					"updated":        Attribute{},
					"title":          Attribute{},
					"content":        Attribute{},
					"author":         Attribute{},
					"labels":         Attribute{},
					"customMetaData": Attribute{},
					"location":       Attribute{},
					"images":         Attribute{},
				},
			},
			"author": View{
				Description: "Author and admin level detail",
				Attributes: Attributes{
					"kind":           Attribute{},
					"id":             Attribute{},
					"published":      Attribute{},
					"url":            Attribute{},
					"updated":        Attribute{},
					"title":          Attribute{},
					"content":        Attribute{},
					"author":         Attribute{},
					"labels":         Attribute{},
					"customMetaData": Attribute{},
					"location":       Attribute{},
					"images":         Attribute{},
					"status":         Attribute{},
				},
			},
		},
	},
}
