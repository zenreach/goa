package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// The sequence of actions is as follows:
// 1. Parse code: Invoke go parser on selected files.
// 2. Process AST: build resources, media types and controller definitions from AST.
// 3. Generate code: process each resource and its dependencies in alphabetical order.
func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fail(err.Error())
	}
	pathFlag := flag.String("path", cwd, "Path to files containing controllers and resources")
	exclude := flag.String("ex", "", "Pattern of files to exclude, you may only use one of \"-ex\" or \"-in\".\n    (see http://golang.org/src/path/match.go?s=1103:1161#L28 for pattern syntax)")
	include := flag.String("in", "", "Pattern of files to include, you may only use one of \"-ex\" or \"-in\".\n    (see http://golang.org/src/path/match.go?s=1103:1161#L28 for pattern syntax)")
	verbose := flag.Bool("verbose", false, "Turn verbose mode on")
	flag.Parse()
	if *exclude != "" && *include != "" {
		fail("You may only specify -in or -ex, not both.")
	}

	// 1. Parse code
	path := *pathFlag
	fset := token.NewFileSet()
	mode := parser.ParseComments
	if *verbose {
		mode += parser.Trace
	}
	packages, err := parser.ParseDir(fset, path, filter(*exclude, *include),
		mode)
	if err != nil {
		fail(err.Error())
	}

	// 2. Analyze AST
	a := newAnalyzer(packages, *verbose)
	api, errs := a.analyze()
	if len(errs) > 0 {
		fail(errs.Error())
	}
	err = api.validate()
	if err != nil {
		fail(err.Error())
	}

	// 3. Generate code
	dest := filepath.Join(path, "codegen.go")
	w, err := os.Create(dest)
	if err != nil {
		fail(fmt.Sprintf("Could not open %s (%s)", dest, err.Error()))
	}
	errs = generateApi(api, w)
	if len(errs) > 0 {
		fail(errs.Error())
	}

	// We're done!
	fmt.Println(dest)
}

// Helper function used to filter source files to be parsed according to the
// 'ex' and 'in' flags.
func filter(exclude, include string) func(os.FileInfo) bool {
	return func(info os.FileInfo) bool {
		if info.Name() == "codegen.go" {
			return false
		}
		if exclude != "" {
			matched, err := filepath.Match(exclude, info.Name())
			if err != nil {
				fail("Failed to load files: " + err.Error())
			}
			if matched {
				return false
			}
		}
		if include != "" {
			matched, err := filepath.Match(include, info.Name())
			if err != nil {
				fail("Failed to load files: " + err.Error())
			}
			if !matched {
				return false
			}
		}
		return true
	}
}

// Print error message and quit process with non-0 return value
func fail(msg string) {
	fmt.Printf("** %s\n", msg)
	os.Exit(1)
}
