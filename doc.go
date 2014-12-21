// goa provides a novel way to build RESTful APIs using go, it uses the same design/implementation separation principle
// introduced by RightScale's praxis framework (http://www.praxis-framework.io).
//
// In goa, API controllers are paired with resource definitions that provide the metadata needed to do automatic
// validation of requests and reponses as well document generation. Resource definitions list each controller action
// describing their parameters, payload and responses.
//
// On top of validation, goa can also use resource definitions to coerce incoming request payloads to the right data
// type (string "true" to boolean true, string "1" to int 1 etc.) alleviating the need for writing all the boilerplate
// code validation and coercion usually require.
//
// goa also provides the following benefits:
// - built-in support for bulk actions using multi-part mime
// - integration with negroni to leverage existing middlewares
// - built-in support for form encoded, multipart form encoded and JSON request bodies
//
// Controllers in goa can be of any type. They simply implement the functions corresponding to the actions defined
// in the resource definition.  These functions take a single argument which implements the Request interface. This
// interface provides access to the request parameter and payload attributes. It also provides the mean to record the
// response to be sent back to the client. The following is a valid goa controller and corresponding resource
// definition:
//
//    var echoResource = Resource{                                           // Resource definition
//       Actions: Actions{                                                   // List of supported actions
//          "echo": Action{                                                  // Only one action "echo"
//             Route: GET("?value={value}"),                                 // Capture param in "value"
//             Params: Params{"value": Param{Type: String, Required: true}}, // Param is a string and must be provided
//             Responses: Responses{"ok": Response{Status: 200}},            // Only one possible response for this action
//          },
//       },
//    }
//
//    type EchoController struct{}                        // EchoController type
//    func (e* echoController) Echo(r Request) {          // Implementation of its "echo" action
//       r.RespondWithBody("ok", r.ParamString("value"))  // Simply response with content of "value" query string
//    }
//
// Once a resource and the corresponding controller are implemented they can be mounted onto a goa application.
// Mounting a controller defines its path. Taking the example above, the following runs the goa app:
//
//    // Launch goa app
//    func main() {
//       app := goa.NewApplication("/echo")          // Create application
//       app.Mount(&echoResource, &EchoController{}) // Mount resource and corresponding controller
//       http.ListenAndServe(":80", app)             // Application implements standard http.Handlefunc
//    }
//
// Given the code above clients may send HTTP requests to `/echo?value=xxx`. The response will have status code 200
// and the body will contain the content of the "value" query string (xxx).
// If the client does not specify the "value" query string then goa automatically generates a response with code 400 and
// a message in the body explaining that the query string is required.
// The resource definition could specify additional constraints on the "value" parameter (e.g. minimum and/or maximum
// length or regular expression) and goa would perform the validation and return 400 responses with clear error messages
// if it failed.
//
// This automatic validation and the document generation (tbd) provide the means for API designers to provide an API
// definition complete with request and response definitions without having to actually implement any code. Future
// changes to the APIs can also be reviewed by simply tweaking the resource definitions with no need to touch
// controller code. This also means the API documentation is always up-to-date.
//
// A note about the goa source code:
// The code is intented to be clear and well documented to make it possible for anyone to browse through and understand
// how the library fits together. The "examples" directory contains a couple of simple examples to help get started.
// Additional more complex examples are in the works.
package goa
