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
	TaskPayload = Object{
		"Owner":     TaskMediaType.Object["Owner"],
		"Details":   TaskMediaType.Object["Details"],
		"Kind":      TaskMediaType.Object["Kind"],
		"ExpiresAt": TaskMediaType.Object["ExpiresAt"]}

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
	User = Object{
		"FirstName": Member(String, "User first name"),
		"LastName":  Member(String, "User last name"),
		"Email":     Member(String, "User email").Format("email"),
	}
	// Media type object (JSON schema definition)
	taskObject := Object{
		"Id":        Member(Integer, "Task id").Minimum(1),
		"Href":      Member(String, "Task href"),
		"Owner":     Member(User, "Task owner"),
		"Details":   Member(String, "Task details").MinLength(1),
		"Kind":      Member(String, "Todo or reminder").Enum("todo", "reminder"),
		"ExpiresAt": Member(String, "Todo expiration or reminder trigger").Format("date-time"),
		"CreatedAt": Member(String, "Task creation timestamp").Format("date-time"),
	}
	task := NewMediaType("application/vnd.acme.task",
		"Task media type, supports a 'tiny' view for quick indexing.",
		taskObject)

	task.Link("CreatedBy").Using("Owner")

	// Views available to render media type
	task.View("Default").As("Id", "Href", "Owner:tiny", "Kind", "ExpiresAt", "CreatedAt").Links("CreatedBy") // Syntax is "PropertyName[:ViewName]
	task.View("Tiny").As("Id", "Href", "Kind", "ExpiresAt").Links("CreatedBy")

	return task
}

// Resource not found response media type, default media type for actions that respond with 404.
func resourceNotFoundMediaType() *MediaType {
	notFoundObject := Object{
		"Id":       Member(Integer, "Id of looked up task").Minimum(1),
		"Resource": Member(String, "Type of looked up resource, e.g. 'tasks'"),
	}

	return NewMediaType("application/vnd.acme.task-not-found",
		"Task not found media type", notFoundObject)
}
