#!/bin/bash -e

# List tasks
echo GET /tasks
echo ----------
curl -i -X GET http://localhost:8081/tasks
echo
echo

# Show one task
echo GET /tasks/1
echo ------------
curl -i -X GET http://localhost:8081/tasks/1
echo
echo

# Create a task
echo POST /tasks
echo -----------
curl -i -X POST http://localhost:8081/tasks -H Content-Type:application/json \
	-d '{"Owner":{"Name":"Joe"},"Details":"a task","ExpiresAt":"2015-03-26T15:02:01-07:00"}'
echo
echo

# Update a task
echo PUT /tasks/1
echo ------------
curl -i -X PUT http://localhost:8081/tasks/1 -H Content-Type:application/json \
	-d '{"Details":"a task","ExpiresAt":"2015-03-26T15:02:01-07:00"}'

# Delete a task
echo DELETE /tasks/2
echo ---------------
curl -i -X DELETE http://localhost:8081/tasks/2
echo
