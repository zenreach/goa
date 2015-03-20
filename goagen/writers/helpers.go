package writers

import (
	"fmt"
	"os"
	"strings"

	"github.com/kr/text"
)

// Common header for all generated files
const genHeader = `//************************************************************************//
// TITLE
//
// Generated with:
{{comment commandLine}}
//
//        The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//
`

// Produce line comments by concatenating given strings and producing 80 characters long lines
// starting with "//"
func comment(elems ...string) string {
	var lines []string
	for _, e := range elems {
		lines = append(lines, strings.Split(e, "\n")...)
	}
	var trimmed = make([]string, len(lines))
	for i, l := range lines {
		trimmed[i] = strings.TrimLeft(l, " \t")
	}
	t := strings.Join(trimmed, "\n")
	return text.Indent(t, "// ")
}

// Command line used to run tool
func commandLine() string {
	return fmt.Sprintf("$ %s %s", os.Args[0], strings.Join(os.Args[1:], " "))
}

// Produces header template using given header title
func header(title string) string {
	return strings.Replace(genHeader, "TITLE", title, 1)
}
