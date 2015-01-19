.PHONY: test clean

goa:
	go build

test: *_test.go
	go test
