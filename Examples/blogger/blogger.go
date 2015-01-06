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
	// Setup application
	app := goa.New("/api")
	app.Mount(&v3.blogController{}, &v3.blogResource)
	app.Mount(&v3.postController{}, &v3.postResource)
	app.Mount(&v3.commentController{}, &v3.commentResource) 

	// Print routes and run app
	l := log.New(os.Stdout, "[blogger] ", 0)
	addr := "localhost:8082"
	l.Printf("listening on %s", addr)
	l.Printf("routes:")
	app.Routes().Log(l)
	l.Printf("\n")
	l.Printf("---------------------------------------------------------------------------------")
	l.Printf("  index   `curl http://%s/api/blogs`", addr)
	l.Printf("  show    `curl http://%s/api/blogs/1`", addr)
	l.Printf("  create: `curl -X POST -d '{\"name\":\"My Blog\"}'\\\n" +
                 "                   -H 'Content-Type:application/json' http://%s/api/blogs`", addr)
	l.Printf("  update: `curl -X PUT -d '{\"name\":\"My Blog 2\"}'\\\n" +
	         "                   -H 'Content-Type:application/json' http://%s/api/blogs/1`", addr)
	l.Printf("  delete: `curl -X DELETE http://%s/api/blogs/1`", addr)
	l.Printf("---------------------------------------------------------------------------------")

	l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
}
