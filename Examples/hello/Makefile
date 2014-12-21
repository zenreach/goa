.PHONY: run routes clean

run:
	godep go run hello.go

routes:
	godep go run hello.go --routes

hello: hello.go pkg/*.go
	godep go build hello.go

clean:
	rm hello
