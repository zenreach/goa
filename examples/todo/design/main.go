package design

import (
	"regexp"

	. "github.com/raphael/goa/design"
)

var (
	// Task resource
	TaskResource *Resource
	// Task resource media type
	TaskMediaType *MediaType
	// Task index media type
	TaskIndexMediaType *MediaType
	// Task not found response media type
	TaskNotFoundMediaType *MediaType
	// Task creation and update request payload
	TaskPayload Object
	// Payload 'User'
	User Object
)

func Main() {
	// Define media types
	TaskMediaType = taskMediaType()
	TaskNotFoundMediaType = resourceNotFoundMediaType()

	// Define request payloads
	TaskPayload = NewObject(
		TaskMediaType.Object["Owner"].Required(),
		TaskMediaType.Object["Details"].Required(),
		TaskMediaType.Object["Kind"].Required(),
		TaskMediaType.Object["ExpiresAt"].Required())

	// Define task resource
	TaskResource = NewResource("Task", "/tasks", "Todo task", TaskMediaType)
	TaskResource.Version = "1.0"

	// Define task actions
	var index = TaskResource.Index("")
	TaskIndexMediaType = index.Responses[0].MediaType

	var show = TaskResource.Show(":id")
	show.WithParam("view") // .String() is implicit
	show.Respond(TaskNotFoundMediaType).WithStatus(404)

	var create = TaskResource.Create("").WithPayload(TaskPayload)
	create.RespondNoContent().WithLocation(regexp.MustCompile(`/tasks/[0-9]+$`))

	var patch = TaskResource.Patch(":id").WithPayload(TaskPayload)
	patch.Respond(TaskNotFoundMediaType).WithStatus(404)

	var del = TaskResource.Delete(":id")
	del.Respond(TaskNotFoundMediaType).WithStatus(404)
}

// Resource media type, default media type for actions that respond with 200.
// Validation rules are inherited by action response media types and action
// request payloads that define attributes with identical names.
// Validations follow the json schema draft 4 validation document -
// http://json-schema.org/latest/json-schema-validation.html.
func taskMediaType() *MediaType {
	// User object
	User = NewObject(
		Prop("FirstName", String, "User first name"),
		Prop("LastName", String, "User last name"),
		Prop("Email", String, "User email"))

	// Media type object (JSON schema definition)
	taskObject := NewObject(
		Prop("Id", Integer, "Task id").Minimum(1),
		Prop("Href", String, "Task href"),
		Prop("Owner", User, "Task owner"),
		Prop("Details", String, "Task details").MinLength(1),
		Prop("Kind", String, "Todo or reminder").Enum("todo", "reminder"),
		Prop("ExpiresAt", String, "Todo expiration or reminder trigger").Format("date-time"),
		Prop("CreatedAt", String, "Task creation timestamp").Format("date-time"))

	task := NewMediaType("application/vnd.acme.task",
		"Task media type, supports a 'tiny' view for quick indexing.",
		taskObject)

	// Views available to render media type
	task.AddView("tiny", "Id", "User:tiny", "Kind", "ExpiresAt") // Syntax is "PropertyName[:ViewName]"

	return task
}

// Resource not found response media type, default media type for actions that respond with 404.
func resourceNotFoundMediaType() *MediaType {
	notFoundObject := NewObject(
		Prop("Id", Integer, "Id of looked up task").Minimum(1),
		Prop("Resource", String, "Type of looked up resource, e.g. 'tasks'"),
	)

	return NewMediaType("application/vnd.acme.task-not-found",
		"Task not found media type", notFoundObject)
}
