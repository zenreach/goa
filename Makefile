.PHONY: doc

test: *.go
	godep go test

doc: test
	godocdown . > ./README.md
	sed '2 a [![GoDoc](https://godoc.org/github.com/raphael/goa?status.svg)](https://godoc.org/github.com/raphael/goa)\
		[![Build Status](https://travis-ci.org/raphael/goa.svg)](https://travis-ci.org/raphael/goa)' ./README.md > ./README.md.2
	mv ./README.md.2 ./README.md



