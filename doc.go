// goa provides a novel way to build RESTful APIs using go, it uses the same design/implementation separation principle
// introduced by RightScale's praxis framework (http://www.praxis-framework.io).
//
// In goa, API controllers are paired with resource definitions that list each action describing their parameters,
// payload and responses. These descriptions are used by goa to automatically validate incoming requests, outgoing
// responses and produce documentation.
// Action definitions leverage media types to describe the responses sent back to clients. Media types are used to
// create documentation that describe the response contents.
//
// Automatic validation means that by the time the controller action is called, all the parameters (parsed from the
// request path) and payload attributes (parsed from the request body) are of the expected type.
//
// Automatic documentation makes it possible to design or change an exiting api and review the changes before writing
// any action implementation code.
//
// goa also provides the following benefits:
// - built-in support for bulk actions using multi-part mime
// - integration with negroni to leverage existing middlewares
// - built-in support for form encoded, multipart form encoded and JSON request bodies
//
// Controllers in goa can be of any type. They must implement functions with the same name as the action name defined
// in the resource definition. These functions take a single argument which implements the Request interface. This
// interface provides access to the request parameter and payload attributes. It also provides the mean to record the
// response to be sent back to the client, see the Request interface for more information.
//
// Resource definitions describe a resource media type and actions. Each action describe the parameters and payload
// attributes as well as the possible responses. goa will validate both the incoming requests and the controller
// response against the definition. See the Resource type for more information.
//
// Once a resource and the correspondin controller are implemented they can be mounted onto the goa application. Mounting
// a controller defines its path. See the Mount function.
//
// The code is intented to be well documented so that it's possible to more easily browse through it and understand how
// the library fits together. The "examples" directory contains a couple of simple examples to help get started. More
// feature-ful examples are being worked on.
package goa
