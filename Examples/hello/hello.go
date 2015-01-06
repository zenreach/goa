// hello world example
// Shows basic usage of goa
//
// Run:
//   make run
// Index:
//   curl http://localhost:8080/api/hello
// Show:
//  curl http://localhost:8080/api/hello/0
// Update:
//   curl -X PUT -d '{"Value":"foo"}' http://localhost:8080/api/hello/1 -H Content-Type:application/json
// Delete:
//   curl -X DELETE http://localhost:8080/api/hello/1
package main

import (
	"./pkg"
	"github.com/raphael/goa"
	"log"
	"net/http"
	"os"
)

func main() {
	// Setup application
	app := goa.New("/api")
	app.Mount(&hello.Hello{}, &hello.HelloResource)

	// Print routes and run app
	l := log.New(os.Stdout, "[hello] ", 0)
	addr := "localhost:8081"
	l.Printf("Listening on %s", addr)
	l.Println("Routes:")
	app.Routes().Log(l)
	l.Printf("\n")
	l.Printf("---------------------------------------------------------------------------------")
	l.Printf("  index   `curl http://%s/api/hello`", addr)
	l.Printf("  show    `curl http://%s/api/hello/1`", addr)
	l.Printf("  create: `curl -X POST -d '{\"value\":\"foo\"}'\\\n" +
                 "                   -H 'Content-Type:application/json' http://%s/api/hello`", addr)
	l.Printf("  update: `curl -X PUT -d '{\"value\":\"foo\"}'\\\n" +
	         "                   -H 'Content-Type:application/json' http://%s/api/hello/1`", addr)
	l.Printf("  delete: `curl -X DELETE http://%s/api/hello/1`", addr)
	l.Printf("---------------------------------------------------------------------------------")

	l.Fatal(http.ListenAndServe(addr, app)) // Application implements standard http.Handlefunc
}
