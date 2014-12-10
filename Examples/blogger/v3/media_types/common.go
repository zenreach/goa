package v3

import (
	. "github.com/raphael/goa"
)

var authorType = Composite{
	"id": Attribute{
		Description: "The author ID.",
		Type:        String,
	},
	"displayName": Attribute{
		Description: "The author display name.",
		Type:        String,
	},
	"url": Attribute{
		Description: "The URL to the author profile page.",
		Type:        String,
	},
	"image": Attribute{
		Description: "The author avatar image.",
		Type:        imageType,
	},
}

var imageType = Composite{
	"url": Attribute{
		Description: "The URL of the avatar image",
		Type:        String,
	},
}

var inReplyToType = Composite{
	"id": Attribute{
		Description: "The ID of the parent of this resource.",
		Type:        String,
	},
}

var repliesType = Composite{
	"totalItems": Attribute{
		Description: "The total number of comments on this post.",
		Type:        Integer,
	},
	"items": Attribute{
		Description: "The list of comments for this post.",
		Type:        CollectionOf(commentMediaType),
	},
	"selfLink": Attribute{
		Description: "The Blogger API URL of to retrieve the comments for this post.",
		Type:        String,
	},
}

var locationType = Composite{
	"name": Attribute{
		Description: "Location name.",
		Type:        String,
	},
	"lat": Attribute{
		Description: "Location's latitude.",
		Type:        Float,
	},
	"lng": Attribute{
		Description: "Location's longitude.",
		Type:        Float,
	},
	"span": Attribute{
		Description: "Location's viewport span. Can be used when rendering a map preview.",
		Type:        String,
	},
}
