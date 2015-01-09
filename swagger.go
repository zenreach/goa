// Package swagger allows generating Swagger documentation for goa apps.
// See http://swagger.io/ for details.
// Documentation for the structs defined in this package can be found at
// https://github.com/swagger-api/swagger-spec/blob/master/versions/2.0.md
// Not all structs are currently used to generate the documentation.
package goa

import (
	"encoding/json"
	"strconv"
)

// This is the root document object for the API specification.
type SwaggerSpec struct {
	Swagger             string                       `json:"swagger"`
	Info                *SwaggerInfo                 `json:"info"`
	Host                string                       `json:"host,omitempty"`
	BasePath            string                       `json:"basePath,omitempty"`
	Schemes             []string                     `json:"schemes,omitempty"`
	Consumes            []string                     `json:"consumes,omitempty"`
	Produces            []string                     `json:"produces,omitempty"`
	Paths               *SwaggerPaths                `json:"paths,omitempty"`
	Definitions         *SwaggerDefinitions          `json:"definitions,omitempty"`
	Parameters          *SwaggerParameterDefinitions `json:"parameters,omitempty"`
	Responses           *SwaggerResponseDefinitions  `json:responses,omitempty`
	SecurityDefinitions *SwaggerSecurityDefinitions  `json:securityDefinitions,omitempty`
	Security            *SwaggerSecurity             `json:security,omitempty`
	Tags                *SwaggerTag                  `json:"tags,omitempty"`
	ExternalDocs        *SwaggerExternalDocs         `json:"externalDocs,omitempty"`
}

// API Info struct
type SwaggerInfo struct {
	Title          string          `json:"title,omitempty"`
	Description    string          `json:"description,omitempty"`
	TermsOfService string          `json:"termsOfService,omitempty"`
	Contact        *SwaggerContact `json:"contact,omitempty"`
	License        *SwaggerLicense `json:"license,omitempty"`
	Version        string          `json:"version,omitempty"`
}

// An object to hold responses to be reused across operations.
type SwaggerResponseDefinitions map[string]SwaggerResponse

// An object to hold parameters to be reused across operations.
type SwaggerParameterDefinitions map[string]SwaggerParameter

// Api contact information
type SwaggerContact struct {
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// API license information
type SwaggerLicense struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

// The Schema Object allows the definition of input and output data types.
type SwaggerSchema struct {
	SwaggerValidated
	Ref           string                   `json:"$ref,omitempty"`
	Format        string                   `json:"format,omitempty"`
	Title         string                   `json:"title,omitempty"`
	Description   string                   `json:"description,omitempty"`
	MaxProperties *int                     `json:"maxProperties,omitempty"`
	MinProperties *int                     `json:"minProperties,omitempty"`
	Required      []interface{}            `json:"required,omitempty"`
	Type          string                   `json:"type,omitempty"`
	Items         *SwaggerItems            `json:"items,omitempty"`
	Properties    map[string]SwaggerSchema `json:"items,omitempty"`
	AllOf         []*SwaggerSchema         `json:"allOf,omitempty"`
	Descriminator string                   `json:"descriminator,omitempty"`
	ReadOnly      *bool                    `json:"readOnly,omitempty"`
	Xml           *SwaggerXml              `json:"xml,omitempty"`
	ExternalDocs  *SwaggerExternalDocs     `json:"externalDocs,omitempty"`
	Example       interface{}              `json:"example,omitempty"`
}

// Holds the relative paths to the individual endpoints.
type SwaggerPaths map[string]SwaggerPathItem

// Describes the operations available on a single path.
type SwaggerPathItem struct {
	Ref        string            `json:"$ref,omitempty"`
	Get        *SwaggerOperation `json:"get,omitempty"`
	Put        *SwaggerOperation `json:"put,omitempty"`
	Post       *SwaggerOperation `json:"post,omitempty"`
	Delete     *SwaggerOperation `json:"delete,omitempty"`
	Options    *SwaggerOperation `json:"options,omitempty"`
	Head       *SwaggerOperation `json:"head,omitempty"`
	Patch      *SwaggerOperation `json:"patch,omitempty"`
	Parameters *SwaggerParameter `json:"parameters,omitempty"`
}

// Describes a single API operation on a path.
type SwaggerOperation struct {
	Tags         []string             `json:"tags,omitempty"`
	Summary      string               `json:"summary,omitempty"`
	Description  string               `json:"description,omitempty"`
	ExternalDocs *SwaggerExternalDocs `json:"externalDocs,omitempty"`
	OperationId  string               `json:"operationId,omitempty"`
	Consumes     []string             `json:"consumes,omitempty"`
	Produces     []string             `json:"produces,omitempty"`
	Parameters   []SwaggerParameter   `json:"parameters,omitempty"`
	Responses    *SwaggerResponses    `json:"responses,omitempty"`
	Schemes      string               `json:"schemes,omitempty"`
	Deprecated   bool                 `json:"deprecated,omitempty"`
	Security     []SwaggerSecurity    `json:"security,omitempty"`
}

// Allows referencing an external resource for extended documentation.
type SwaggerExternalDocs struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

// Describes a single operation parameter.
type SwaggerParameter struct {
	SwaggerValidated
	Ref         string         `json:"$ref,omitempty"`
	Name        string         `json:"name,omitempty"`
	In          string         `json:"in,omitempty"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      *SwaggerSchema `json:"schema,omitempty"`
	Type        string         `json:"type,omitempty"`
	Format      string         `json:"format,omitempty"`
	Items       *SwaggerItems  `json:"items,omitempty"`
}

// Used by parameter definitions that are not located in "body".
type SwaggerItems struct {
	SwaggerValidated
	Type             string        `json:"type,omitempty"`
	Format           string        `json:"format,omitempty"`
	Items            *SwaggerItems `json:"items,omitempty"`
	CollectionFormat string        `json:"collectionFormat,omitempty"`
}

// A container for the expected responses of an operation.
type SwaggerResponses map[string]SwaggerResponse

// Describes a single response from an API Operation.
type SwaggerResponse struct {
	Description string          `json:"description,omitempty"`
	Schema      *SwaggerSchema  `json:"schema,omitempty"`
	Headers     *SwaggerHeaders `json:"headers,omitempty"`
	Examples    *SwaggerExample `json:"examples,omitempty"`
}

// Lists the headers that can be sent as part of a response.
type SwaggerHeaders map[string]SwaggerHeader

// Describes a single response header from an API Operation.
type SwaggerHeader struct {
	SwaggerValidated
	Description      string `json:"description,omitempty"`
	Type             string `json:"type,omitempty"`
	Format           string `json:"format,omitempty"`
	CollectionFormat string `json:"collectionFormat,omitempty"`
}

// Allows sharing examples for operation responses.
type SwaggerExample map[string]interface{}

// Allows adding meta data to a single tag that is used by the Operation Object.
type SwaggerTag struct {
	Name         string               `json:"name,omitempty"`
	Description  string               `json:"description,omitempty"`
	ExternalDocs *SwaggerExternalDocs `json:"externalDocs,omitempty"`
}

// A metadata object that allows for more fine-tuned XML model definitions.
type SwaggerXml struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute *bool  `json:"attribute,omitempty"`
	Wrapped   *bool  `json:"wrapped,omitempty"`
}

// An object to hold data types that can be consumed and produced by operations.
type SwaggerDefinitions map[string]SwaggerSchema

// An object to hold parameters to be reused across operations.
type SwaggerParameters map[string]SwaggerParameter

// A declaration of the security schemes available to be used in the
// specification.
type SwaggerSecurityDefinitions map[string]SwaggerSecurityScheme

// Allows the definition of a security scheme that can be used by the operations.
type SwaggerSecurityScheme struct {
	Type             string         `json:"type,omitempty"`
	Description      string         `json:"description,omitempty"`
	Name             string         `json:"name,omitempty"`
	In               string         `json:"in,omitempty"`
	Flow             string         `json:"flow,omitempty"`
	AuthorizationUrl string         `json:"authorizationUrl,omitempty"`
	TokenUrl         string         `json:"tokenUrl,omitempty"`
	Scopes           *SwaggerScopes `json:"scopes,omitempty"`
}

// Lists the available scopes for an OAuth2 security scheme.
type SwaggerScopes map[string]string

// Lists the required security schemes to execute this operation.
type SwaggerSecurity map[string]string

// Validated data structure
type SwaggerValidated struct {
	Default          interface{}   `json:"default,omitempty"`
	Maximum          *int          `json:"maximum,omitempty"`
	ExclusiveMaximum *bool         `json:"exclusiveMaximum,omitempty"`
	Minimum          *int          `json:"minimum,omitempty"`
	ExclusiveMinimum *bool         `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int          `json:"maxLength,omitempty"`
	MinLength        *int          `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int          `json:"maxItems,omitempty"`
	MinItems         *int          `json:"minItems,omitempty"`
	UniqueItems      *bool         `json:"uniqueItems,omitempty"`
	Enum             []interface{} `json:"enum,omitempty,omitempty"`
	MultipleOf       *int          `json:"multipleOf,omitempty"`
}

// Generate produces a swagger spec object from a goa app
func GenerateSwagger(ap Application, info *SwaggerInfo, host string) string {
	paths := SwaggerPaths{}
	a := ap.(*app)
	for _, r := range a.resources {
		for _, a := range r.actions {
			for _, route := range a.routes {
				p := SwaggerPathItem{}
				op := operation(r, a, route)
				switch route.verb {
				case "GET":
					p.Get = op
				case "HEAD":
					p.Head = op
				case "POST":
					p.Post = op
				case "PUT":
					p.Put = op
				case "DELETE":
					p.Delete = op
				case "PATCH":
					p.Patch = op
				case "OPTIONS":
					p.Options = op
				}
				paths[route.path] = p
			}
		}
	}

	spec := SwaggerSpec{
		Swagger:  "2.0",
		Info:     info,
		Host:     host,
		BasePath: "/", // Actions specify full path
		Paths:    &paths,
	}

	if res, err := json.Marshal(spec); err != nil {
		panic("goa: failed to generate swagger docs - " + err.Error())
	} else {
		return string(res)
	}

}

// operation generates describes a single API Operation.
func operation(r *compiledResource, a *compiledAction, route *compiledRoute) *SwaggerOperation {
	responses := SwaggerResponses{}
	produces := []string{}
	for _, resp := range a.responses {
		var schema *SwaggerSchema
		if md := resp.mediaType; md != nil {
			produces = append(produces, md.Identifier)
			schema = schemaFromModel(&md.Model)
		}
		st := resp.response.Status
		if st == 0 {
			st = 200
		}
		responses[strconv.Itoa(st)] = SwaggerResponse{
			Description: resp.response.Description,
			Schema:      schema,
			Headers:     headersFromResponse(resp),
		}
	}
	consumes := []string{}
	if a.payload != nil {
		consumes = []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data"}
	}
	description := r.description
	if len(description) > 0 {
		description += "\n"
	}
	description += a.name
	if len(a.description) > 0 {
		description += "\n" + a.description
	}
	return &SwaggerOperation{
		Tags:        []string{"resource=" + r.name},
		Summary:     a.description,
		Description: description,
		OperationId: r.name + "." + a.name,
		Consumes:    consumes,
		Produces:    produces,
		Parameters:  parameters(r, a),
		Responses:   &responses,
	}
}

func schemaFromModel(m *Model) *SwaggerSchema {
	return nil // TBD
}

func headersFromResponse(r *compiledResponse) *SwaggerHeaders {
	return nil // TBD
}

func parameters(r *compiledResource, a *compiledAction) []SwaggerParameter {
	return nil // TBD
}
