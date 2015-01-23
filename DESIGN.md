Public Interface
----------------
goa.RegisterControllers
goa.ServeHTTP

goa.MountHandler
goa.WriteRaml

Generated
---------
mountAllHandlers

Handler:
- Validate and parse params & payload
- Create controller with r, w
- controller exposes Respond()
- Calls controller action
- Validates response
- Serializes response (can be overridden
- Sends response

