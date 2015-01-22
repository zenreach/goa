package goa

import (
	"reflect"
)

// Media types are used to define the content of controller action responses.
// They provide the API clients with a crisp definition of what an API returns.
// Conceptually media types define the "views" of resources, the media type
// content is described using a json schema.
//
// A media type also define views which provide different ways of rendering its
// properties. For example there could be a "tiny" view that only includes a few
// attributes (e.g. "name", "href") used for listing and an "extended" view that
// includes all the properties.
//
// Finally, a media type may define a `links` property to represent related
// resources. The links property type must be an object, each property
// corresponding to a link. The type of the links properties must be reference
// to media types that define a "link" view. The link view is a special view
// that is used when rendering links to the media type defining it.
//
// A media type may define both a property and a link with the same name, views
// may then decide to use one or another (or even both). When a link is defined
// with the same name as an attribute of the media type then it does not have to
// redefine any of the attribute field, they get "inherited" from the media type
// attribute.
type MediaType struct { 
	Identifier  string     // HTTP media type identifier (http://en.wikipedia.org/wiki/Internet_media_type)
	Description string     // Description used for documentation
	Schema      JsonSchema // Actual media type definition
	Views       Views      // Media type views
}

// Views have a description and attributes
type View struct {
	Description string   // View description
	Properties  []string // Name of properties to include in view
}

// Collection of named Views
type Views map[string]View
