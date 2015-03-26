#!/bin/bash -e

# List tasks
curl -X GET http://localhost:8081/tasks
echo

# Show one task
curl -X GET http://localhost:8081/tasks/1
echo

# Create a task
curl -X POST http://localhost:8081/tasks -H Content-Type:application/json -d "{}"
echo

# Update a task
curl -X PUT http://localhost:8081/tasks/1 -H Content-Type:application/json -d "{}"
echo

# Delete a task
curl -X DELETE http://localhost:8081/tasks/1
echo
