package main

import (
	"./v3/controllers"
	"./v3/resources"
	"github.com/raphael/goa"
	"net/http"
)

func main() {
	app := goa.NewApplication("/api")
	app.Mount(&postResource, &controllers.postController{})

	http.handle("/", app.ServeHttp)
}
