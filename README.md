# goa
--
    import "github.com/raphael/goa"

[![GoDoc](https://godoc.org/github.com/raphael/goa?status.svg)](https://godoc.org/github.com/raphael/goa) [![Build Status](https://travis-ci.org/raphael/goa.svg)](https://travis-ci.org/raphael/goa)

goa provides a novel way to build RESTful APIs using go, it uses the same
design/implementation separation principle introduced by RightScale's praxis
framework (http://www.praxis-framework.io).

In goa, API controllers are paired with resource definitions that provide the
metadata needed to do automatic validation of requests and reponses as well as
document generation. Resource definitions list each controller action describing
their parameters, payload and responses.

On top of validation, goa also uses resource definitions to coerce incoming
request payloads to the right data type (e.g. string "true" to boolean true,
string "1" to int 1 etc.) alleviating the need for writing all the boilerplate
code that validation and coercion usually require.

goa also provides the following benefits:

    - built-in support for bulk actions using multi-part mime
    - integration with Negroni (http://godoc.org/github.com/codegangsta/negroni)
      to leverage existing middlewares
    - built-in support for form encoded, multipart form encoded and JSON request
      bodies

A controller in goa can be any object that exposes methods corresponding to the
actions defined in the resource definition. The first argument of the action
methods is alway a goa request object. If the action definition specifies a
payload for the ation (i.e. if the corresponding HTTP requests have a non-empty
body) then the second argument of action methods is a pointer to an instance of
the payload blueprint struct. The rest of the arguments contain the values of
the action parameters (URL captures and query strings). The request object
passed in the first argument exposes methods that let actions specify the
content of the HTTP response.

The following is a valid goa resource definition:

    // Echo resource
    var EchoResource = Resource{                                           // Resource definition
       Actions: Actions{                                                   // List of supported actions
          "echo": Action{                                                  // Only one action "echo"
             Route: GET("?value={value}"),                                 // Capture param in "value"
             Params: Params{"value": Param{Type: String, Required: true}}, // Query string "value" is a string and must be provided
             Responses: Responses{"ok": Http.Ok()},                        // Only one possible response for this action
          },
       },
    }

And the following a controller that implements it:

    // Echo controller
    type EchoController struct {}

    // Implementation of "echo" action
    func (e* echoController) Echo(r *Request, value string) {
       r.Respond(value)
    }

Controller are mounted onto a goa application using the `Mount()` method. Taking
the example above, the following runs the goa app:

    // Launch goa app
    func main() {
       app := goa.New("/echo")                     // Create application
       app.Mount(&EchoController{}, &EchoResource) // Mount resource and corresponding controller
       http.ListenAndServe(":80", app)             // Application implements standard http.Handlefunc
    }

Given the code above clients may send HTTP requests to `/echo?value=xxx`. The
response has a status code of 200 and the body contains the value of the query
string named "value" (xxx). If the client does not specify the "value" query
string then goa returns a response with code 400 and a message in the body
explaining that the query string is required. The resource definition could
specify additional constraints on the "value" parameter (e.g. minimum and/or
maximum length or regular expression) and goa would perform the validation and
return 400 responses with clear error messages if it failed.

This automatic validation and the document generation (tbd) provide the means
for API designers to create an API definition complete with request and response
definitions without having to actually implement any code. Future changes to the
APIs can also be reviewed by simply tweaking the resource definitions with no
need to touch controller code. This also means the API documentation is always
up-to-date.

A note about the goa source code: The code is intented to be clear and well
documented to make it possible for anyone to browse through and understand how
the library fits together. The "examples" directory contains a couple of simple
examples to help you get started. Additional more complex examples are in the
works.
