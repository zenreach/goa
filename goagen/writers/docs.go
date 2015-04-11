package writers

import (
	"bytes"
	"fmt"
	"text/template"

	"gopkg.in/alecthomas/kingpin.v1"
)

// Doc writer.
type docsGenWriter struct {
	designPkg string
	tmpl      *template.Template
}

// Create middleware writer.
func NewDocsGenWriter(pkg string) (Writer, error) {
	tmpl, err := template.New("docs-gen").Parse(docsGenTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &docsGenWriter{designPkg: pkg, tmpl: tmpl}, nil
}

func (w *docsGenWriter) FunctionName() string {
	return "genDocs"
}

func (w *docsGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.tmpl.Execute(&buf, w), "docs-gen template")
	return buf.String()
}

const docsGenTmpl = `
func {{.FunctionName}}(resource *design.Resource, output string) error {
	return nil
}`
