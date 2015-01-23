package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	pathFlag := flag.String("path", ".", "Path to files containing controllers and resources")
	exclude := flag.String("ex", "", "Pattern of files to exclude, you may only use one of \"-ex\" or \"-in\".\n    (see http://golang.org/src/path/match.go?s=1103:1161#L28 for pattern syntax)")
	include := flag.String("in", "", "Pattern of files to include, you may only use one of \"-ex\" or \"-in\".\n    (see http://golang.org/src/path/match.go?s=1103:1161#L28 for pattern syntax)")
	flag.Parse()
	if exclude != nil && include != nil {
		fail("You may only specify -in or -ex, not both.")
	}
	var path string
	if pathFlag == nil {
		path, _ = os.Getwd()
	} else {
		path = *pathFlag
	}
	fset := token.NewFileSet()
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
	packages, err := parser.ParseDir(fset, path, filter(exclude, include), parser.ParseComments)
	if err != nil {
		fail(err.Error())
	}
	analyze(packages)
}

// Helper function used to filter source files to be parsed according to the
// 'ex' and 'in' flags.
func filter(exclude, include *string) func(os.FileInfo) bool {
	return func(info os.FileInfo) bool {
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

// Extract resources and controllers
func analyze(packages map[string]*ast.Package) {
	description = newServiceDescription()
	errors := []error{}
	for _, pkg := range packages {
		for name, object := range pkg.Scope.Objects {
			if object.Kind != ast.Typ {
				continue
			}
			spec := object.Decl.(*ast.TypeSpec)
			docs := spec.Doc.List
			header := *(docs[0])
			if len(docs) < 15 { // len("//@goa Resource") == 15
				continue
			}
			if header[:7] != "//@goa " {
				continue
			}
			switch header[8:] {
			case "Resource":
				res, err := parseResource(spec)
				if err != nil {
					errors := append(errors, fmt.Errorf("Failed to parse resource %s: %s", spec.Name.Name, err.Error()))
				} else {
					resources[name] = resources
				}
			case "MediaType":
				md, err := parseMediaType(spec)
				if err != nil {
					errors := append(errors, fmt.Errorf("Failed to parse media type %s: %s", spec.Name.Name, err.Error()))
				} else {
					mediaTypes[name] = md
				}
			case "Controller":
				ctr, err := parseController(spec)
				if err != nil {
					errors := append(errors, fmt.Errorf("Failed to parse controller %s: %s", spec.Name.Name, err.Error()))
				} else {
					controllers[name] = ctr
				}
			default:
				errors := append(errors, fmt.Errorf("Unknown @goa directive '%s', directive must be one of 'MediaType', 'Resource' or 'Controller'"))
			}
		}
	}
}

// TBD: Check that action parameters use JSON compatible types (numbers, bool or string)
func parseResource(spec *ast.TypeSpec) (*resource, error) {
	for d := range spec.Doc.List {
		comment := *d
		if len(comment) < 8 {
			continue
		}
		if comment[:7] != "//@goa " {
			continue
		}
	}
	return &resource{}, nil
}

func parseMediaType(spec *ast.TypeSpec) (*mediaType, error) {
	return &mediaType{}, nil
}

func parseController(spec *ast.TypeSpec) (*controller, error) {
	return &controller{}, nil
}

// Print error message and quit process with non-0 return value
func fail(msg string) {
	fmt.Printf("** %s\n", msg)
	os.Exit(1)
}
