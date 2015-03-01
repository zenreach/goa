package main

func TaskResource() *goa.Resource {
	// Define task
	var task = goa.NewResource("Task", "/tasks", NewTaskMediaType())
	task.Description("Todo task")
	task.Version("1.0")

	// Common response media types
	var taskNotFound = NewTaskNotFoundMediaType()

	// Define task actions
	var show = task.Show("id")             // same as task.Action("show").Get(":id")
	show.Description("Get a task")         // optional, goa generates default descriptions for CRUD actions
	show.Param(":id").Integer()            // can chain other validations, like .Required(), .Pattern(), .MinLength() etc.
	show.Respond(taskNotFound).Status(404) // show.Respond(TaskMediaType()) is implied (because Show) also .Status(200) is default

	task.Index("")

	var create = task.Action("create").Post("")
	create.Payload(NewCreateTaskPayload())
	create.RespondNoContent().Header("Location") // .Status(204) is implied

	var patch = task.Patch(":id")
	patch.Param(":id").Integer() 
	patch.Payload(NewPatchTaskPayload())
	patch.Respond(taskNotFound).Status(404)

	var delete = task.Delete(":id")
	delete.Param(":id").Integer()
	delete.Respond(taskNotFound).Status(404)
}

// Resource media type, default media type for actions that respond with 200.
// Validation rules are inherited by action response media types and action
// request payloads that define attributes with identical names.
// Validations follow the json schema draft 4 validation document -
// http://json-schema.org/latest/json-schema-validation.html.
func NewTaskMediaType() *goa.MediaType {
	task = goa.NewMediaType("application/vnd.acme.task")

	// Optional description
	task.Description("Task media type, supports a 'tiny' view for quick indexing.")

	// Media type attributes
	task.Attr("Id", "Task id").Integer().Mininum(1)
	task.Attr("Href", "Task href").String()
	task.Attr("Owner", "Task owner").Type("User")
	task.Attr("Details", "Task detils").String().MinLength(1)
	task.Attr("Kind", "Todo or reminder").String().Enum([]string{"todo", "reminder"})
	task.Attr("ExpiresAt", "Todo expiration or reminder trigger").String().Format("date-time")
	task.Attr("CreatedAt", "Task creation timestamp").String().Format("date-time")

	// Views available to render media type
	task.DefaultView("Id", "User", "Details", "Kind", "ExpiresAt", "CreatedAt") // optional, default view includes all attributes
	task.View("tiny", "Id", "User:tiny", "Kind", "ExpiresAt")                   // Syntax is "AttributeName[:ViewName]"

	return task
}