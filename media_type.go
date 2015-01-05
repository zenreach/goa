package goa

import (
	"reflect"
)

// Media types are used to define the content of controller action responses. They provide the API clients with a crisp
// definition of what an API returns. Conceptually media types define the "views" of resources, the media type content is
// described using goa types (see attribute.go). The media type definition also defines the go data structure used to
// handle the corresponding data in the application, a valid definition of a media type could be:
//
//     // Article data structure
//     type Article struct {
//         Href    string // API href
//         Title   string // Title
//         Content string // Content
//     }
//
//     // Show blog article response media type
//     articleMediaType := MediaType{
//
//        Identifier: "application/vnd.blogapp.article+json",
//
//        Description: "A blog article",
//
//        Model: Model {
//            Blueprint: Article{},
//
//            Attributes: Attributes{    // An article has an href, a title and a content
//                "href": Attribute{
//                    Type:        String,
//                    Description: "Article href",
//                    MinLength:   4,
//                },
//                "title": Attribute{
//                    Type:        String,
//                    Description: "Article title",
//                    MinLength:   20,
//                    MaxLength:   100,
//                },
//                "content": Attribute{
//                    Type:        String,
//                    Description: "Article content",
//                    MinLength:   100,
//                 },
//            },
//        },
//
//     }
//
// Media types implement the `Type` interface so that they may be used when specifying the type of an attribute. As an
// example this makes it possible to define an API where a resource may be retrieved by itself or as part of another
// (e.g. parent) resource.
//
// A media type also define views which provide different ways of rendering its attributes. For example there could be a
// "tiny" view that only includes a few attributes (e.g. name, href) used for listing and an "extended" view that
// includes all the attributes.
//
// Finally, a media type may define a `links` attribute to represent related resources. The links attribute type must be
// `Composite`, each field corresponding to a link. The type of the links attribute fields must be media types themselves
// that define a "link" view. The link view is a special view that is used when rendering links to the media type
// defining it.
//
// A media type may define both an attribute and a link with the same name, views may then decide to use one or another
// (or even both). When a link is defined with the same name as an attribute of the media type then it does not have to
// redefine any of the attribute field, they get "inherited" from the media type attribute (see example below).
//
// Extending the example above with links:
//
//     // Article media type
//     articleMediaType := MediaType{
//
//        Identifier: "application/vnd.blogapp.article+json",
//
//        Description: "A blog article",
//
//        Model: Model{
//            Blueprint: Article{},
//
//            Attributes:{  // An article has an href, a title and a content
//                "href": Attribute{
//                    Type:        String,
//                    Description: "Article href",
//                    MinLength:    4,
//                },
//                "title": Attribute{
//                    Type:        String,
//                    Description: "Article title",
//                    MinLength:   20,
//                    MaxLength:   100,
//                },
//                "content": Attribute{
//                    Type:        String,
//                    Description: "Article content",
//                    MinLength:   100,
//                },
//            },
//        },
//
//        Views: Views{    // An article can be linked to (defines a "link" view)
//            "default": View{
//                Description: "default view",
//                Attributes:    Attributes{
//                    "href":    Attribute{},
//                    "title":   Attribute{},
//                    "content": Attribute{},
//                },
//            },
//            "link": View{
//                Description: "href only",
//                Attributes: Attributes{
//                    "href": Attribute{}
//                },
//            },
//        },
//     }
//
//     // The blog media type contains a collection of articles, the default view contains links to articles and the
//     // extended view embeds the article.
//
//     type Blog struct {
//         Name     string
//         Articles []*Article
//     }
//
//     var blogMediaType = MediaType{
//
//        Identifier: "application/vnd.blogapp.blog+json",
//
//        Description: "A blog",
//
//        Model: {
//            Blueprint: Blog{},
//
//            Attributes: Attributes{    // A blog has a name and articles
//                "name": Attribute{
//                    Type:        String,
//                    Description: "Blog name",
//                },
//                "articles": Attribute{
//                    Type:        CollectionOf(articleMediaType),
//                    Description: "Blog articles",
//                },
//                "links": Attributes{        // A blog has a "articles" link
//                    "articles": Attribute{} // No need to redefine the "articles" attribute
//                },
//            },
//        },
//
//        Views: Views{
//            "default": View{    // The default view contains the blog name and links to articles
//                Description: "default view",
//                Attributes:  Attributes{
//                    "name":  Attribute{},
//                    "links": Attribute{},
//                },
//            },
//            "extended": View{    // The extended view embeds the articles
//                Description: "extended view",
//                Attributes: Attributes{
//                    "name":     Attribute{},
//                    "articles": Attribute{},
//                },
//            },
//        },
//
//    }
type MediaType struct {
	Identifier  string // HTTP media type identifier (http://en.wikipedia.org/wiki/Internet_media_type)
	Description string // Description used for documentation
	Model       Model  // Actual media type definition
	Views       Views  // Media type views
}

// Views have a description and attributes
type View struct {
	Description string     // View description
	Attributes  Attributes // Attributes to include in view, can override attribute fields from media type
}

// IsEmpty returns true if media type is empty (does not have an identifier, attributes or views), false otherwis
func (m *MediaType) IsEmpty() bool {
	return len(m.Identifier) == 0 && len(m.Model.Attributes) == 0 && len(m.Views) == 0
}

// Collection of named Views
type Views map[string]View

// Media types implement the `Type` interface so they can be used interchangeably where types are (e.g. in attribute
// definitions).

// Load load the given value into an instance of the blueprint struct.
// `value` must either be a map with string keys or a string containing a JSON representation of a map.
// Load also applies any validation rule defined in the media type model attributes.
// Returns nil and an error if coercion or validation fails.
func (m *MediaType) Load(value interface{}) (interface{}, error) {
	return m.Model.Load(value)
}

// CanLoad checks whether values of the given go type can be loaded into values of this media type model.
// Returns nil if check is successful, error otherwise.
func (m *MediaType) CanLoad(t reflect.Type, context string) error {
	c := Composite(m.Model.Attributes)
	return c.CanLoad(t, context)
}

// GetKind returns the kind of this type (media type)
func (m *MediaType) GetKind() Kind {
	return TMediaType
}

// Media type Type kind
const TMediaType Kind = _TLast

// Marker for media type inherited from resource
// Substituted with actual resource media type upon "compilation"
func resourceMediaType() MediaType {
	return MediaType{
		Identifier:  "Resource",
		Description: "Resource media type",
	}
}

// Media type inherited from resource for collection
// Substituted with actual resource collection media type upon "compilation"
func resourceCollectionMediaType() MediaType {
	return MediaType{
		Identifier:  "ResourceCollection",
		Description: "Resource collection media type",
	}
}
