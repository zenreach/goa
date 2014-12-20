.PHONY: run routes clean

run:
	godep go run minimal.go

routes:
	godep go run minimal.go --routes

minimal: minimal.go
	godep go build minimal.go

clean:
	rm minimal
