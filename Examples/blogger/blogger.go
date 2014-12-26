// Blogger goa application
// Showcases most of goa features with a realistic example
// API is strongly inspired from Google's blogger API:
//   https://developers.google.com/blogger/docs/3.0/reference/index
// Application implements 3 controllers to expose the blog, post and comment
// resources.
// The example shows one way of doing versioning using go packages to implement
// different versions.
package main

import (
	"../../"
	"./v3/controllers"
	"./v3/resources"
	"net/http"
)

func main() {
	// Setup --routes flag
	printRoutes := flag.Bool("routes", false, "Print routes")
	flag.Parse()

	// Setup application
	app := goa.NewApplication("/api")
	app.Mount(&v3.blogResource, &v3.blogController{})
	app.Mount(&v3.postResource, &v3.postController{})
	app.Mount(&v3.commentResource, &v3.commentController{})

	// Print routes or run app
	if *printRoutes {
		app.PrintRoutes()
	} else {
		l := log.New(os.Stdout, "[hello] ", 0)
		addr := "localhost:8080"
		l.Printf("listening on %s", addr)
		l.Printf("---------------------------------------------------------------------------------------------------------")
		l.Printf("  index  with `curl http://%s/api/[blogs|posts|comments]`", addr)
		l.Printf("  show   with `curl http://%s/api/[blogs|posts|comments]/1`", addr)
		l.Printf("  update with `curl -X PUT -d '{\"Value\":\"foo\"}' -H Content-Type:application/json http://%s/api/[blogs|posts|comments]/1`", addr)
		l.Printf("  delete with `curl -X DELETE http://%s/api/[blogs|posts|comments]/1`", addr)
		l.Printf("---------------------------------------------------------------------------------------------------------")

		l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
	}
}
