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
		"FirstName": M(String, "User first name"),
		"LastName":  M(String, "User last name"),
		"Email":     M(String, "User email").Format("email"),
	}
	// Media type object (JSON schema definition)
	taskObject := Object{
		"Id":        M(Integer, "Task id").Minimum(1),
		"Href":      M(String, "Task href"),
		"Owner":     M(User, "Task owner"),
		"Details":   M(String, "Task details").MinLength(1),
		"ExpiresAt": M(String, "Todo expiration or reminder trigger").Format("date-time"),
		"CreatedAt": M(String, "Task creation timestamp").Format("date-time"),
	}
	// Actual media type
	task := NewMediaType("application/vnd.acme.task",
		"Task media type, supports a 'tiny' view for quick indexing.",
		taskObject)

	task.Link("Owner").As("CreatedBy")

	// Views available to render media type
	task.View("default").With("Id", "Href", "Owner:tiny", "Details", "ExpiresAt", "CreatedAt").Link("CreatedBy") // Syntax is "MemberName[:ViewName]
	task.View("tiny").With("Id", "Href", "ExpiresAt").Link("CreatedBy")

	return task
}

// Resource not found response media type, default media type for actions that respond with 404.
func resourceNotFoundMediaType() *MediaType {
	notFoundObject := Object{
		"Id":       M(Integer, "Id of looked up task").Minimum(1),
		"Resource": M(String, "Type of looked up resource, e.g. 'tasks'"),
	}

	return NewMediaType("application/vnd.acme.task-not-found",
		"Task not found media type", notFoundObject)
}
