package writers

import (
	"fmt"
	"text/template"
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
	return ""
}

func (w *docsGenWriter) Source() string {
	return ""
}

const docsGenTmpl = ``
