package v3

import (
	"../../models"
	. "github.com/raphael/goa"
)

// Blog user
var blogUserInfo = Composite{
	"kind": Attribute{
		Description: "The kind of this entity. Always goa#blogger#blogUserInfo",
		Type:        String,
		Regexp:      "^goa#blogger#blogUserInfo$",
	},
	"blog": Attribute{
		Description: "The Blog resource.",
		Type:        blogMediaType,
	},
	"blogUserInfo": Attribute{
		Description: "Information about a User for the Blog.",
		Type: Composite{
			"kind": Attribute{
				Description: "The kind of this entity. Always goa#blogger#blogPerUserInfo",
				Type:        String,
			},
			"userId": Attribute{
				Description: "ID of the User",
				Type:        String,
			},
			"blogId": Attribute{
				Description: "ID of the Blog resource",
				Type:        String,
				Regexp:      "[0-9]+",
			},
			"photosAlbumKey": Attribute{
				Description: "The Photo Album Key for the user when adding photos to the blog",
				Type:        String,
			},
			"hasAdminAccess": Attribute{
				Description: "True if the user has Admin level access to the blog.",
				Type:        Boolean,
			},
		},
	},
}

var blogListMediaType = MediaType{

	Identifier: "application/vnd.example.blogger.blogList+json",

	Model: Model{
		Blueprint: greeting{},

		Attributes: Attributes{

			"kind": Attribute{
				Description: "The kind of this entity. Always goa#blogger#blogList",
				Type:        String,
			},

			"blogs": Attribute{
				Description: "The list of Blogs this user has Authorship or Admin rights for.",
				Type:        CollectionOf(blogMediaType),
			},

			"blogUserInfos": Attribute{
				Description: "Admin level list of blog per-user information",
				Type:        CollectionOf(blogUserInfo),
			},
		},
	},
}

/* Blog media type */
var blogMediaType = MediaType{

	Identifier: "vnd.example.blogger.post",

	Model: Model{
		Blueprint: model.Blog{},

		Attributes: Attributes{
			"kind": Attribute{
				Description: "The kind of this resource. Always goa#blogger#blog",
				Type:        String,
			},
			"id": Attribute{
				Description: "The ID for this resource.",
				Type:        String,
				Regexp:      "[0-9]+",
			},
			"name": Attribute{
				Description: "The name of this blog, which is usually displayed in Blogger as the blog's title. The title can include HTML.",
				Type:        String,
			},
			"description": Attribute{
				Description: "The description of this blog, which is usually displayed in Blogger underneath the blog's title. The description can include HTML.",
				Type:        String,
			},
			"published": Attribute{
				Description: "RFC 3339 date-time when this blog was published, for example \"2007-02-09T10:13:10-08:00\".",
				Type:        Time,
			},
			"updated": Attribute{
				Description: "RFC 3339 date-time when this blog was last updated, for example \"2012-04-15T19:38:01-07:00\".",
				Type:        Time,
			},
			"url": Attribute{
				Description: "The URL where this blog is published.",
				Type:        String,
			},
			"selfLink": Attribute{
				Description: "The Blogger API URL to fetch this resource from.",
				Type:        String,
			},
			"posts": Attribute{
				Description: "The container for this blog's posts.",
				Type:        repliesType,
			},
			"locale": Attribute{
				Description: "The location, if this post is geotagged.",
				Type: Composite{
					"language": Attribute{
						Description: "The language this blog is set to, for example \"en\" for English.",
						Type:        String,
					},
					"country": Attribute{
						Description: "The country variant of the language, for example \"US\" for American English.",
						Type:        String,
					},
					"variant": Attribute{
						Description: "The language variant this blog is set to.",
						Type:        String,
					},
				},
			},
			"customMetaData": Attribute{
				Description: "",
				Type:        String,
			},
		},
	},

	Views: Views{
		"default": View{
			Description: "Only view available",
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

		"tiny": View{
			Description: "ID only",
			Attributes: Attributes{
				"id": Attribute{},
			},
		},
	},
}
