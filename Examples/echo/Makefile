.PHONY: run routes clean

run:
	godep go run echo.go

routes:
	godep go run echo.go --routes

echo: echo.go
	godep go build echo.go

clean:
	rm echo
