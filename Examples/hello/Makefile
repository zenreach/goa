.PHONY: run clean

run:
	godep go run hello.go

hello: hello.go pkg/*.go
	godep go build hello.go

clean:
	rm -f hello
