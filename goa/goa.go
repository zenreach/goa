package main

import (
	"flag"
	"fmt"
	"go/parser"
	"os"
	"path/filepath"
	"strings"
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
	flag.Parse()
	if exclude != nil && include != nil {
		fail("You may only specify -in or -ex, not both.")
	}
	path := *pathFlag
	//fset := token.NewFileSet()
	// base := 0
	// filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
	// if len(exclude) > 0 && path.Match(exclude, path) {
	// return nil
	// }
	// if len(include) > 0 && !path.Match(include, path) {
	// return nil
	// }
	// size := info.Size()
	// fset.AddFile(filePath, base, size)
	// base += size
	// return nil
	// })

	// 1. Parse code
	packages, err := parser.ParseDir(nil, path, filter(exclude, include),
		parser.ParseComments)
	if err != nil {
		fail(err.Error())
	}

	// 2. Analyze AST
	api, errs := analyze(packages)
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
func filter(exclude, include *string) func(os.FileInfo) bool {
	return func(info os.FileInfo) bool {
		if !strings.HasSuffix(info.Name(), ".go") {
			return false
		}
		if exclude != nil {
			matched, err := filepath.Match(*exclude, info.Name())
			if err != nil {
				fail("Failed to load files: " + err.Error())
			}
			if matched {
				return false
			}
		}
		if include != nil {
			matched, err := filepath.Match(*include, info.Name())
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
