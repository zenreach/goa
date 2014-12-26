.PHONY: doc

test: *.go
	godep go test

doc: test
	godocdown -output=README.md



