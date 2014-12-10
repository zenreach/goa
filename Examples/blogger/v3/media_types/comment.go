package v3

import (
	. "github.com/raphael/goa"
)

var commentListMediaType = MediaType{

	MimeType: "vnd.example.blogger.commentList",

	Type: Composite{
		"kind": Attribute{
			Description: "The kind of this entity. Always goa#blogger#commentList",
			Type:        String,
		},
		"nextPageToken": Attribute{
			Description: "Pagination token to fetch the next page, if one exists.",
			Type:        String,
		},
		"prevPageToken": Attribute{
			Description: "Pagination token to fetch the previous page, if one exists.",
			Type:        String,
		},
		"items": Attribute{
			Description: "The list of comments resources for the specified post.",
			Type:        CollectionOf(commentMediaType),
		},
	},
}

/* Blog post comment media type */
var commentMediaType = MediaType{

	MimeType: "vnd.example.blog.comment",

	Type: Composite{
		"kind": Attribute{
			Description: "The kind of this resource. Always goa#blogger#comment",
			Type:        String,
		},
		"id": Attribute{
			Description: "The ID for this resource.",
			Type:        String,
			Regexp:      "[0-9]+",
		},
		"post": Attribute{
			Description: "Data about the post containing this comment.",
			Type:        postMediaType,
			View:        "tiny",
		},
		"blog": Attribute{
			Description: "Data about the blog containing this comment.",
			Type:        blogMediaType,
			View:        "tiny",
		},
		"published": Attribute{
			Description: "date-time when this comment was published",
			Type:        Datetime,
		},
		"updated": Attribute{
			Description: "date-time when this comment was last updated",
			Type:        Datetime,
		},
		"content": Attribute{
			Description: "The content of the comment, which can include HTML markup.",
			Type:        String,
			MinLength:   1,
			MaxLength:   65000,
		},
		"author": Attribute{
			Description: "Comment author",
			Type:        authorType,
		},
		"inReplyTo": Attribute{
			Description: "Data about the comment this is in reply to.",
			Type:        inReplyToType,
		},
		"status": Attribute{
			Description: "The status of the comment. The status is only visible to users who have Administration rights on a blog.",
			Type:        String,
			AllowedValues: Values{
				"emptied": "Comments that have had their content removed",
				"live":    "Comments that are publicly visible",
				"pending": "Comments that are awaiting administrator approval",
				"spam":    "Comments marked as spam by the administrator",
			},
		},
	},
}

// Media Type returned by index request
var commentListMediaType = MediaType{

	MimeType: "vnd.example.blog.commentList",

	Type: Composite{
		"kind": Attribute{
			Description: "The kind of this resource. Always goa#blogger#commentList",
			Type:        String,
		},
		"nextPageToken": Attribute{
			Description: "Pagination token to fetch the next page, if one exists.",
			Type:        String,
		},
		"prevPageToken": Attribute{
			Description: "Pagination token to fetch the previous page, if one exists.",
			Type:        String,
		},
		"items": Attribute{
			Description: "The list of comments resources for the specified post.",
			Type:        CollectionOf(commentMediaType),
		},
	},

	Views: Views{
		View{Name: "reader",
			Description: "Reader level detail (default)",
			Attrs:       AttRefs{"kind", "id", "published", "updated", "content", "author", "inReplyTo"},
		},
		View{Name: "author",
			Description: "Author level detail (default)",
			Attrs:       AttRefs{"kind", "id", "published", "updated", "content", "author", "inReplyTo", "status"},
		},
		View{Name: "admin",
			Description: "Admin level detail (default)",
			Attrs:       AttRefs{"kind", "id", "published", "updated", "content", "author", "inReplyTo", "status"},
		},
	},
}
