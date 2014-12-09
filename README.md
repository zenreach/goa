# goa
--
    import "github.com/raphael/goa"

goa provides a novel way to build RESTful APIs using go, it uses the same
design/implementation separation principle introduced by RightScale's praxis
framework (http://www.praxis-framework.io).

In goa, API controllers are paired with resource definitions that provide the
metadata needed to do automatic validation of requests and reponses as well
document generation. Resource definitions list each controller action describing
their parameters, payload and responses.

On top of validation, goa can also use resource definitions to coerce incoming
request payloads to the right data type (string "true" to boolean true, string
"1" to int 1 etc.) alleviating the need for writing all the boilerplate code
validation and coercion usually require.

goa also provides the following benefits: - built-in support for bulk actions
using multi-part mime - integration with negroni to leverage existing
middlewares - built-in support for form encoded, multipart form encoded and JSON
request bodies

Controllers in goa can be of any type. They simply implement the functions
corresponding to the actions defined in the resource definition. These functions
take a single argument which implements the Request interface. This interface
provides access to the request parameter and payload attributes. It also
provides the mean to record the response to be sent back to the client. The
following is a valid goa controller and corresponding resource definition:

    var echoResource = Resource{                                           // Resource definition
       Actions: Actions{                                                   // List of supported actions
          "echo": Action{                                                  // Only one action "echo"
             Route: GET("?value={value}"),                                 // Capture param in "value"
             Params: Params{"value": Param{Type: String, Required: true}}, // Param is a string and must be provided
             Responses: Responses{"ok": Response{Status: 200}},            // Only one possible response for this action
          },
       },
    }

    type EchoController struct{}                        // EchoController type
    func (e* echoController) Echo(r Request) {          // Implementation of its "echo" action
       r.RespondWithBody("ok", r.ParamString("value"))  // Simply response with content of "value" query string
    }

Once a resource and the corresponding controller are implemented they can be
mounted onto a goa application. Mounting a controller defines its path. Taking
the example above, the following runs the goa app:

    // Launch goa app
    func main() {
       app := goa.NewApplication("/echo")          // Create application
       app.Mount(&echoResource, &EchoController{}) // Mount resource and corresponding controller
       http.ListenAndServe(":80", app)             // Application implements standard http.Handlefunc
    }

Given the code above clients may send HTTP requests to "/echo?value=xxx". The
response will have status code 200 and the body will contain the content of the
"value" query string (xxx). If the client does not specify the "value" query
string then goa automatically generates a response with code 400 and a message
in the body explaining that the query string is required. The resource
definition could specify additional constraints on the "value" parameter (e.g.
minimum and/or maximum length or regular expression) and goa would perform the
validation and return 400 responses with clear error messages if it failed.

This automatic validation and the document generation (tbd) provide the means
for API designers to provide an API definition complete with request and
response definitions without having to actually implement any code. Future
changes to the APIs can also be reviewed by simply tweaking the resource
definitions with no need to touch controller code. This also means the API
documentation is always up-to-date.

A note about the goa source code: The code is intented to be clear and well
documented to make it possible for anyone to browse through and understand how
the library fits together. The "examples" directory contains a couple of simple
examples to help get started. Additional more complex examples are in the works.

## Usage

```go
const (
	SupportsMultipart = iota // Action request body may use multipart content type
	RequiresMultipart        // Action request body must use multipart content type
)
```
Possible values for the Action struct "Multipart" field

```go
var Boolean = basic(TBoolean)
```
Boolean basic type

```go
var Float = basic(TFloat)
```
Float basic type

```go
var Http f
```
HTTP Response factory

```go
var Integer = basic(TInteger)
```
Integer basic type

```go
var String = basic(TString)
```
String basic type

```go
var Time = basic(TTime)
```
Time basic type

#### func  Multi

```go
func Multi(routes ...singleRoute) multiRoute
```
Multi creates a multi-route from the given list of routes

#### type Action

```go
type Action struct {
	Name        string
	Description string
	Route       Route
	Params      Attributes
	Payload     Attributes
	Filters     Attributes
	Views       []string
	Responses   Responses
	Multipart   int
}
```

Action definitions define a route which consists of one ore more pairs of HTTP
verb and path. They also optionally define the action parameters (variables
defined in the route path) and payload (request body content). Parameters and
payload are described using attributes which may include validations. goa takes
care of validating and coercing the parameters and payload fields (and returns a
response with status code 400 and a description of the validation error in the
body in case of failure).

The Multipart field specifies whether the request body must(RequiresMultipart)
or can (SupportsMultipart) use a multipart content type. Multipart requests can
be used to implement bulk actions - for example bulk updates. Each part contains
the payload for a single resource, the same payload that would be used to apply
the action to that resource in a standard (non-multipart) request.

Action definitions may also specify a list of supported filters - for example an
index action may support filtering the list of results given resource field
values. Filters are defined using attributes, they are specified by the client
using the special "filters" URL query string, the syntax is:

    "?filters[]=some_field==some_value&&filters[]=other_field==other_value"

Filters are readily available to the action implementation after they have been
validated and coerced by goa. The exact semantic is up to the action
implementation.

Action definitions also specify the set of views supported by the action.
Different views may render the media type differently (ommitting certain
attributes or links, see media_type.go). As with filters the client specifies
the view in the special "view" URL query string:

    "?view=tiny"

Finally, action definitions describe the set of potential responses they may
return and for each response the status code, compulsory headers and a media
type (if different from the resource media type). These response definitions are
named so that the action implementation can create a response from its
definition name.

#### func (*Action) ValidateResponse

```go
func (a *Action) ValidateResponse(data ResponseData) error
```
ValidateResponse checks that the response content matches one of the action
response definitions if any

#### type Actions

```go
type Actions map[string]Action
```

Map of action definitions keyed by action name

#### type Application

```go
type Application interface {
	// Mount a controller
	Mount(definition *Resource, controller Controller)
	// Goa apps implement the standard http.HandlerFunc
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	// PrintRoutes prints application routes to stdout
	PrintRoutes()
}
```

Public interface of a goa application

#### func  NewApplication

```go
func NewApplication(basePath string) Application
```
Create new goa application given a base path

#### type ArgumentError

```go
type ArgumentError interface {
	Error() string         // Error message
	Stack() string         // Error stack trace
	ArgName() string       // Name of invalid argument
	ArgValue() interface{} // Value of invalid argument
}
```

Argument error

#### func  NewArgumentError

```go
func NewArgumentError(msg string, argName string, argValue interface{}) ArgumentError
```
Build argument error from message, argument name and value

#### type Attribute

```go
type Attribute struct {
	Type         Type        // Attribute type
	Description  string      // Attribute description
	DefaultValue interface{} // Attribute default value (if any), underlying (go) type is dictated by `Type`

	// - Validation rules -
	Required      bool          // Whether the attribute is required when loading a value of this type
	Regexp        string        // Regular expression used to validate string values
	MinValue      interface{}   // Minimum value used to validate integer, float and time values
	MaxValue      interface{}   // Maximum value used to validate integer, float and time values
	MinLength     int           // Minimum value length used to validate strings and collections
	MaxLength     int           // Maximum value length used to validate strings and collections
	AllowedValues []interface{} // White list of possible values, underlying type is dictated by Type
}
```

Attributes are used to describe data structures: an attribute defines a field
name and type. It allows providing a description as well as type-specific
validation rules (regular expression for strings, min and max values for
integers etc.).

The type of a field defined by an attribute may be one of 5 basic types (string,
integer, float, boolean or time) or may be a composite type: another data
structure also defined with attributes. A type may also be a collection (in
which case it defines the type of the elements) or a hash (in which case it
defines the type of the values, keys are always strings).

All types (basic, composite, collection and hash) implement the `Type`
interface. This interface exposes the `Load()` function which accepts the JSON
representation of a value as well as other compatible representations (e.g.
`int8`, `int16`, etc. for integers, "1", "true" for booleans etc.). This makes
it possible to call it recursively on embedded data structures and coerce all
fields to the type defined by the attributes (or fail if the value cannot be
coerced). The Type interface also exposes a `CanLoad()` function which takes a
go type (`reflect.Type`) and returns an error if values of that type cannot be
coerced into the goa type or nil otherwise. The idea here is that `CanLoad` may
be used to validate that a go structure matches an attribute definition.
`Load()` can then be called multiple times to load values into instances of that
structure.

Validation rules apply whenever a value is loaded via the `Load()` method. They
specify whether a field is required, regular expressions (for string
attributes), minimum and maximum length (strings and collections) or minimum and
maximum values (for integer, float and time attributes).

Finally, attributes may also define a default value and/or a list of allowed
values for a field.

Here is an example of an attribute definition using a composite type:

    article := Attribute{
        Description: "An article",
        Type: Composite{
            "title": Attribute{
                Type:        String,
                Description: "Article title",
                MinLength:   20,
                MaxLength:   200,
                Required:    true,
            },
            "author": Attribute{
                Type: Composite{
                    "firstName": Attribute{
                        Type:        String,
                        Description: "Author first name",
                    },
                    "lastName": Attribute{
                        Type:        String,
                        Description: "Author last name",
                    },
                },
                Required: true,
            },
            "published": Attribute{
                Type:        Time,
                Description: "Article publication date",
                Required:    true,
            },
        },
    }

The example above could represent values such as:

    articleData := map[string]interface{}{
        "title": "goa, a novel go web application framework",
        "author": map[string]interface{}{
            "firstName": "Leeroy",
            "lastName":  "Jenkins",
        },
        "published": time.Now(),
    }

#### func (*Attribute) Validate

```go
func (a *Attribute) Validate() error
```
Validate checks that the given attribute struct is properly initialized

#### type Attributes

```go
type Attributes map[string]Attribute
```

Attributes map

#### type Collection

```go
type Collection struct{ ElemType Type }
```

Collection type

#### func (*Collection) CanLoad

```go
func (c *Collection) CanLoad(t reflect.Type, context string) error
```
CanLoad checks whether values of the given go type can be loaded into values of
this collection type. Returns nil if check is successful, error otherwise.

#### func (*Collection) GetKind

```go
func (c *Collection) GetKind() Kind
```
GetKind returns the kind of this type (collection)

#### func (*Collection) Load

```go
func (c *Collection) Load(value interface{}) (interface{}, error)
```
Load coerces the given value into a []interface{} where the array values have
all been coerced recursively. `value` must either be a slice, an array or a
string containing a JSON representation of an array. Load also applies any
validation rule defined in the collection type element attributes. Returns nil
and an error if coercion or validation fails.

#### type Composite

```go
type Composite Attributes
```

Composite type i.e. attributes map

#### func (Composite) CanLoad

```go
func (c Composite) CanLoad(t reflect.Type, context string) error
```
CanLoad checks whether values of the given go type can be loaded into values of
this composite type. Returns nil if check is successful, error otherwise.

#### func (Composite) GetKind

```go
func (c Composite) GetKind() Kind
```
GetKind returns the kind of this type (composite)

#### func (Composite) Load

```go
func (c Composite) Load(value interface{}) (interface{}, error)
```
Load coerces the given value into a map[string]interface{} where the map values
have all been coerced recursively. `value` must either be a map with string keys
or to a string containing a JSON representation of a map. Load also applies any
validation rule defined in the composite type attributes. Returns `nil` and an
error if coercion or validation fails.

#### type Controller

```go
type Controller interface{}
```

A goa controller can be any type (it just needs to implement one function per
action it exposes)

#### type Error

```go
type Error interface {
	Error() string // Error message
	Stack() string // Error stack trace
}
```

Error with stack trace

#### func  NewError

```go
func NewError(msg string) Error
```
Build error with stack trace information

#### func  NewErrorf

```go
func NewErrorf(format string, a ...interface{}) Error
```
Helper method with fmt.Errorf like behavior

#### type Hash

```go
type Hash struct{ ElemType Type }
```

Hash type

#### func (*Hash) CanLoad

```go
func (h *Hash) CanLoad(t reflect.Type, context string) error
```
CanLoad checks whether values of the given go type can be loaded into values of
this hash type. Returns nil if check is successful, error otherwise.

#### func (*Hash) GetKind

```go
func (h *Hash) GetKind() Kind
```
GetKind returns the kind of this type (hash)

#### func (*Hash) Load

```go
func (h *Hash) Load(value interface{}) (interface{}, error)
```
Load coerces the given value into a map[string]interface{} where the map values
have all been coerced recursively. `value` must either be a map with string keys
or a string containing a JSON representation of a map. Load also applies any
validation rule defined in the hash type element attributes. Returns nil and an
error if coercion or validation fails.

#### type Headers

```go
type Headers map[string]string
```

Response header definitions, map of definitions keyed by header name. There are
two kinds of definitions:

    - Regexp definitions consist of strings starting and ending with slash ("/").
    - Exact matches consist of strings that do not start or do not end with slash (or neither).

All action responses are validated against provided header defintions, at least
one of the response headers must match each definition.

#### type IncompatibleType

```go
type IncompatibleType struct {
}
```

Error raised when a values of given go type cannot be assigned to attribute's
type (by `CanLoad()`)

#### func (*IncompatibleType) Error

```go
func (e *IncompatibleType) Error() string
```
Error returns the error message

#### type IncompatibleValue

```go
type IncompatibleValue struct {
}
```

Error raised when a value cannot be coerced to attribute's type (by `Load()`)

#### func (*IncompatibleValue) Error

```go
func (e *IncompatibleValue) Error() string
```
Error returns the error message

#### type Kind

```go
type Kind int
```

Attribute kind

```go
const (
	//	Kind                   Go type produced by Load()
	TString     Kind = iota // string
	TInteger                // int64
	TFloat                  // float64
	TBoolean                // bool
	TTime                   // time.Time
	TComposite              // map[string]interface{}
	TCollection             // []interface{}
	THash                   // map[string]interface{}

)
```
List of supported kinds

```go
const TMediaType Kind = _TLast
```
Media type Type kind

#### type MediaType

```go
type MediaType struct {
	Identifier  string // HTTP media type identifier (http://en.wikipedia.org/wiki/Internet_media_type)
	Description string // Description used for documentation
	Model       Model  // Actual media type definition
	Views       Views  // Media type views
}
```

Media types are used to define the content of controller action responses. They
provide the API clients with a crisp definition of what an API returns.
Conceptually media types define the "views" of resources, the media type content
is described using goa types (see attribute.go). The media type definition also
defines the go data structure used to handle the corresponding data in the
application, a valid definition of a media type could be:

    // Article data structure
    type Article struct {
        Href    string // API href
        Title   string // Title
        Content string // Content
    }

    // Show blog article response media type
    articleMediaType := MediaType{

       Identifier: "application/vnd.blogapp.article+json",

       Description: "A blog article",

       Model: Model {
           Blueprint: Article{},

           Attributes: Attributes{    // An article has an href, a title and a content
               "href": Attribute{
                   Type:        String,
                   Description: "Article href",
                   MinLength:   4,
               },
               "title": Attribute{
                   Type:        String,
                   Description: "Article title",
                   MinLength:   20,
                   MaxLength:   100,
               },
               "content": Attribute{
                   Type:        String,
                   Description: "Article content",
                   MinLength:   100,
                },
           },
       },

    }

Media types implement the `Type` interface so that they may be used when
specifying the type of an attribute. As an example this makes it possible to
define an API where a resource may be retrieved by itself or as part of another
(e.g. parent) resource.

A media type also define views which provide different ways of rendering its
attributes. For example there could be a "tiny" view that only includes a few
attributes (e.g. name, href) used for listing and an "extended" view that
includes all the attributes.

Finally, a media type may define a `links` attribute to represent related
resources. The links attribute type must be `Composite`, each field
corresponding to a link. The type of the links attribute fields must be media
types themselves that define a "link" view. The link view is a special view that
is used when rendering links to the media type defining it.

A media type may define both an attribute and a link with the same name, views
may then decide to use one or another (or even both). When a link is defined
with the same name as an attribute of the media type then it does not have to
redefine any of the attribute field, they get "inherited" from the media type
attribute (see example below).

Extending the example above with links:

     // Article media type
     articleMediaType := MediaType{

        Identifier: "application/vnd.blogapp.article+json",

        Description: "A blog article",

        Model: Model{
            Blueprint: Article{},

            Attributes:{  // An article has an href, a title and a content
                "href": Attribute{
                    Type:        String,
                    Description: "Article href",
                    MinLength:    4,
                },
                "title": Attribute{
                    Type:        String,
                    Description: "Article title",
                    MinLength:   20,
                    MaxLength:   100,
                },
                "content": Attribute{
                    Type:        String,
                    Description: "Article content",
                    MinLength:   100,
                },
            },
        },

        Views: Views{    // An article can be linked to (defines a "link" view)
            "default": View{
                Description: "default view",
                Attributes:    Attributes{
                    "href":    Attribute{},
                    "title":   Attribute{},
                    "content": Attribute{},
                },
            },
            "link": View{
                Description: "href only",
                Attributes: Attributes{
                    "href": Attribute{}
                },
            },
        },
     }

     // The blog media type contains a collection of articles, the default view contains links to articles and the
     // extended view embeds the article.

     type Blog struct {
         Name     string
         Articles []*Article
     }

     var blogMediaType = MediaType{

        Identifier: "application/vnd.blogapp.blog+json",

        Description: "A blog",

        Model: {
            Blueprint: Blog{},

            Attributes: Attributes{    // A blog has a name and articles
                "name": Attribute{
                    Type:        String,
                    Description: "Blog name",
                },
                "articles": Attribute{
                    Type:        CollectionOf(articleMediaType),
                    Description: "Blog articles",
                },
                "links": Attributes{        // A blog has a "articles" link
                    "articles": Attribute{} // No need to redefine the "articles" attribute
                },
            },
        },

        Views: Views{
            "default": View{    // The default view contains the blog name and links to articles
                Description: "default view",
                Attributes:  Attributes{
                    "name":  Attribute{},
                    "links": Attribute{},
                },
            },
            "extended": View{    // The extended view embeds the articles
                Description: "extended view",
                Attributes: Attributes{
                    "name":     Attribute{},
                    "articles": Attribute{},
                },
            },
        },

    }

#### func (*MediaType) CanLoad

```go
func (m *MediaType) CanLoad(t reflect.Type, context string) error
```
CanLoad checks whether values of the given go type can be loaded into values of
this media type model. Returns nil if check is successful, error otherwise.

#### func (*MediaType) GetKind

```go
func (m *MediaType) GetKind() Kind
```
GetKind returns the kind of this type (media type)

#### func (*MediaType) IsEmpty

```go
func (m *MediaType) IsEmpty() bool
```
IsEmpty returns true if media type is empty (does not have an identifier,
attributes or views), false otherwis

#### func (*MediaType) Load

```go
func (m *MediaType) Load(value interface{}) (interface{}, error)
```
Load load the given value into an instance of the blueprint struct. `value` must
either be a map with string keys or a string containing a JSON representation of
a map. Load also applies any validation rule defined in the media type model
attributes. Returns nil and an error if coercion or validation fails.

#### type Model

```go
type Model struct {
	Attributes Attributes
	Blueprint  interface{}
}
```

Models contain the REST resource data. They can be instantiated from a REST
request payload, from raw database data or any other generic representation
(JSON or maps keyed by field names).

Model definitions describe the model attributes and a "blueprint" which is the
zero value of the go struct used by the business logic of the app.

For example the blueprint of a model definition that contains a person name and
age attributes could be defined as:

    struct { Name string; Age int }{}

This example would require the model attributes to contain a "Name" attribute of
type String and a "Age" attribute of type Integer. A blueprint may also use tags
to specify the corresponding attribute names:

    type person struct {
        FirstName string `attribute:"first_name"`
        LastName  string `attribute:"last_name"`
    }

Given the "person" struct defined above, the blueprint `person{}` can be used to
instantiate a model definition with "first_name" and "last_name" attributes:

    personDefinition := NewModel(
        Attributes{
            "first_name": Attribute{Type: goa.String, MinLength: 1},
            "last_name":  Attribute{Type: goa.String, MinLength: 1},
        },
        person{}
    )

Given the definition above the app may instantiate instances of the blueprint
type using the `Load()` function. The function will take care of validating and
coercing the input data into the struct used by the app to implement the logic.
While model definitions are used internally by the framework to load request
payloads they may also be used independently by the app for example to load data
from a database.

Note that while both model definitions and media types are defined using
attributes the semantic is different: model definition attributes describe the
actual data structure used by the app while media type attributes define how
responses are built.

#### func  NewModel

```go
func NewModel(attributes Attributes, blueprint interface{}) (*Model, error)
```
Create new model definition given named attributes and a blueprint Return an
error if the blueprint is invalid (i.e. not a struct) or if the blueprint fields
do not match the attributes.

#### func (*Model) Load

```go
func (m *Model) Load(value interface{}) (interface{}, error)
```
Load a map indexed by field names or its JSON representation into instance of
model definition blueprint struct. Argument must be either a string (JSON) or a
map whose keys are strings. Returns a pointer to struct with the same type as
the blueprint whose fields have been initialized from given data.

Example:

    // Data structures used by application logic
    type Address struct {
        Street string `attribute:"street"`
        City   string `attribute:"city"`
    }
    type Employee struct {
        Name    string   `attribute:"name"`
        Title   string   `attribute:"title"`
        Address *Address `attribute:"address"`
    }

    // Model definition attributes
    attributes := Attributes{
        "name": Attribute{
            Type:      String,
            MinLength: 1,
            Required:  true,
        },
        "title": Attribute{
            Type:      String,
            MinLength: 1,
            Required:  true,
        },
        "address": Attribute{
            Type: Composite{
                "street": Attribute{
                    Type: String,
                },
                "city": Attribute{
                    Type:      String,
                    MinLength: 1,
                    Required:  true,
                },
            },
        },
    }

    // Create model definition from attributes and blueprint
    definition, _ := NewModel(attributes, Employee{})

    // Data coming from external source (API payload, data store etc.)
    data := map[string]interface{}{
        "name":  "John",
        "title": "Accountant",
        "address": map[string]interface{}{
            "street": "5779 Lamey Drive",
            "city":   "Santa Barbara",
        },
    }

    // Load data into application data structures
    if raw, err := definition.Load(&data); err == nil {
        employee := raw.(*Employee)
        fmt.Printf("Employee: %+v\n", *employee)
    } else {
        fmt.Printf("Load failed: %s\n", err.Error())
    }

#### func (*Model) Validate

```go
func (m *Model) Validate() error
```

#### type Request

```go
type Request interface {
	Respond(response ResponseData)                                        // Send given response
	RespondEmpty(name string)                                             // Helper to respond with empty response
	RespondWithBody(name string, body interface{})                        // Helper to respond with body
	RespondWithHeader(name string, body interface{}, header *http.Header) // Helper to respond with body and headers
	RespondInternalError(body interface{})                                // Helper to respond with 500 and error message

	ResponseBuilder(name string) (ResponseBuilder, error) // Retrieve response builder to build more complex responses

	Param(name string) interface{}   // Retrieve parameter, requires type assertion before value can be used
	ParamString(name string) string  // Retrieve string parameter
	ParamInt(name string) int64      // Retrieve integer parameter
	ParamBool(name string) bool      // Retrieve boolean parameter
	ParamFloat(name string) float64  // Retrieve float parameter
	ParamTime(name string) time.Time // Retrieve time parameter

	Payload(name string) interface{}   // Retrieve payload attribute, requires type assertion before value can be used
	PayloadString(name string) string  // Retrieve string payload attribute
	PayloadInt(name string) int64      // Retrieve integer payload attribute
	PayloadBool(name string) bool      // Retrieve boolean payload attribute
	PayloadFloat(name string) float64  // Retrieve float payload attribute
	PayloadTime(name string) time.Time // Retrieve time payload attribute
}
```

Controller actions take a `Request` interface as only parameter. The interface
exposes methods to retrieve the coerced request parameters and payload
attributes as well as the raw HTTP request object.

The same interface also exposes methods to send the response back. There are a
few ways to do this:

    - use `Respond()` to specify a response object. This object can be any object that implements the Response

interface.

    - use `RespondEmpty()`, `RespondWithBody()` or `RespondWithHeader()` to send a response given its name and content

the name of a response must match one of the action response definition names

    - use `RespondInternalError()` to return an error response (status 500)

The Request interface also exposes a `ResponseBuilder()` method which given a
response name returns an object that can be used to build the corresponding
response (which can then be sent using the `Respond()` method described above)

#### type Resource

```go
type Resource struct {
	Description string
	ApiVersion  string
	RoutePrefix string
	MediaType   MediaType
	Actions     map[string]Action
}
```

Resource definitions describe REST resources exposed by the application API.
They can be versioned so that multiple versions can be exposed (usually for
backwards compatibility). Clients specify the version they want to use through
the X-API-VERSION request header. If an api version is specified in the resource
definition then clients must specify the version header or get back a response
with status code 404.

The definition also includes a description, a route prefix, a media type and
action definitions. The route prefix is the common path under which all resource
actions are located. The complete URL for an action is application path +
controller prefix + resource prefix + action path. The media type describes the
fields of the resource (see media_type.go). The resource actions may define a
different media type for their responses, typically the "index" and "show"
actions re-use the resource media type. Action definitions list all the actions
supported by the resource (both CRUD and other actions), see the Action struct.

#### type Response

```go
type Response struct {
	Description string    // Description used by documentation
	Status      int       // Response status code
	MediaType   MediaType // Response media type if any
	Location    string    // Response 'Location' header validation, enclose value in / for regexp behavior
	Headers     Headers   // Response header validations, enclose values in / for regexp behavior
	Parts       *Response // Response part definitions if any
}
```

Response definitions dictate the set of valid responses a given action may
return. A response definition describes the response status code, media type and
compulsory headers. The 'Location' header is called out as it is a common header
returned by actions that create resources A multipart response definition may
also describe compulsory headers for its parts.

#### func (*Response) NewResponse

```go
func (d *Response) NewResponse() ResponseData
```
Factory method to create corresponding responses

#### func (*Response) Validate

```go
func (d *Response) Validate(r ResponseData) error
```
Validate checks response against definition. Returns error if validation fails,
nil otherwise.

#### type ResponseBuilder

```go
type ResponseBuilder interface {
	SetHeader(name, value string)
	AddHeader(name, value string)
	SetBody(body string)
	AddPart(part ResponseData)
	Response() ResponseData
}
```

The ResponseBuilder interface exposes methods use by actions to initialize the
response.

#### type ResponseData

```go
type ResponseData interface {
	Status() int                    // HTTP response status
	Header() *http.Header           // HTTP response headers
	Body() interface{}              // HTTP response body
	Parts() map[string]ResponseData // Multipart response parts if any
	PartId() string                 // Multipart response inner part id if any
}
```

ResponseData provides access to the HTTP response data. goa provides a default
implementation and various factory methods for building the response. Actions
may alternatively initialize the Response field of the request object they
receive as argument with their own implementation of the interface.

#### type Responses

```go
type Responses map[string]Response
```

Map of response definitions keyed by response name

#### type Route

```go
type Route interface {
	GetRawRoutes() [][]string // Retrieve pair of HTTP verb and action path
}
```

Interface implemented by action route

#### func  DELETE

```go
func DELETE(path string) Route
```
DELETE creates a route with DELETE verb and given path

#### func  GET

```go
func GET(path string) Route
```
GET creates a route with GET verb and given path

#### func  HEAD

```go
func HEAD(path string) Route
```
HEAD creates a route with HEAD verb and given path

#### func  OPTIONS

```go
func OPTIONS(path string) Route
```
OPTIONS creates a route with OPTIONS verb and given path

#### func  PATCH

```go
func PATCH(path string) Route
```
PATCH creates a route with PATCH verb and given path

#### func  POST

```go
func POST(path string) Route
```
POST creates a route with POST verb and given path

#### func  PUT

```go
func PUT(path string) Route
```
PUT creates a route with PUT verb and given path

#### func  TRACE

```go
func TRACE(path string) Route
```
TRACE creates a route with TRACE verb and given path

#### type RouteMap

```go
type RouteMap []*routeData
```

The RouteMap type exposes two public methods WriteRoutes and PrintRoutes that
can be called to print the routes for all mounted resource actions.

#### func (*RouteMap) PrintRoutes

```go
func (m *RouteMap) PrintRoutes()
```
PrintRoutes prints routes to stdout

#### func (*RouteMap) WriteRoutes

```go
func (m *RouteMap) WriteRoutes(writer io.Writer)
```
WriteRoutes writes routes table to given io writer

#### type Type

```go
type Type interface {
	GetKind() Kind                                // Type kind, one of constants defined above
	Load(value interface{}) (interface{}, error)  // Load value, return error if `CanLoad()` would or a validation fails
	CanLoad(t reflect.Type, context string) error // nil if values of given type can be loaded into fields described by attribute, descriptive error otherwise
}
```

Interface implemented by all types (basic, composite, collection and hash)

#### func  CollectionOf

```go
func CollectionOf(t Type) Type
```
CollectionOf creates a collection (array) type. Takes type of elements as
argument.

#### func  HashOf

```go
func HashOf(t Type) Type
```
HashOf creates a hash type. Takes type of keys and values as argument, hash keys
are always strings.

#### type View

```go
type View struct {
	Description string     // View description
	Attributes  Attributes // Attributes to include in view, can override attribute fields from media type
}
```

Views have a description and attributes

#### type Views

```go
type Views map[string]View
```

Collection of named Views
