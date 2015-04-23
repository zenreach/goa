// This package contains the generator writers. These writer produce the go code for the
// generators which in turn produce go code for the application handlers, middleware, client etc.
// The goagen utility instantiates the writers corresponding to the given command line options and
// use their methods to produce code that is then compiled into the application that produces the
// final artefacts.
package writers
