package resources

import . "github.com/raphael/goa/design"

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
	TaskPayload *Member
	// Payload 'User'
	User Object
)

func Init() {
	// Define media types
	TaskMediaType = taskMediaType()
	TaskNotFoundMediaType = resourceNotFoundMediaType()

	// Define request payloads by reusing media type object members
	taskPayloadType := Object{
		"Owner":     TaskMediaType.Object["Owner"],
		"Details":   TaskMediaType.Object["Details"],
		"ExpiresAt": TaskMediaType.Object["ExpiresAt"],
	}
	TaskPayload := M(taskPayloadType, "Task creation and update payload").
		Required("Details", "ExpiresAt")

	// Define task resource
	TaskResource = NewResource("Task", "/tasks", "Todo task", TaskMediaType)
	TaskResource.Version = "1.0"

	// Define task resource actions
	var index = TaskResource.Index("")
	TaskIndexMediaType = index.Responses[0].MediaType

	var show = TaskResource.Show(":id")
	show.WithParam("view")
	show.Respond(TaskNotFoundMediaType).WithStatus(404)

	TaskResource.Create("").WithPayload(TaskPayload)

	var update = TaskResource.Update(":id").WithPayload(TaskPayload)
	update.Respond(TaskNotFoundMediaType).WithStatus(404)

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
	}
	// Actual media type
	task := NewMediaType(
		"application/vnd.acme.task",
		"Task media type, supports a 'tiny' view for quick indexing.",
		taskObject,
	)
	task.Link("Owner").As("CreatedBy")

	// Views available to render media type
	task.View("default").With("Id", "Href", "Owner:tiny", "Details", "ExpiresAt").Link("CreatedBy") // Syntax is "MemberName[:ViewName]
	task.View("tiny").With("Id", "Href", "ExpiresAt").Link("CreatedBy")

	return task
}

// Resource not found response media type, default media type for actions that respond with 404.
func resourceNotFoundMediaType() *MediaType {
	return NewMediaType(
		"application/vnd.acme.task-not-found",
		"Task not found media type",
		Object{
			"Id":       M(Integer, "Id of looked up task").Minimum(1),
			"Resource": M(String, "Type of looked up resource, e.g. 'tasks'"),
		},
	)
}
