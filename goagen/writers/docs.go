package writers

import (
	"fmt"
	"text/template"
)

// Doc writer.
type docsWriter struct {
	designPkg string
	tmpl      *template.Template
}

// Create middleware writer.
func NewDocsWriter(pkg string) (Writer, error) {
	funcMap := template.FuncMap{"joinNames": joinNames, "literal": literal}
	tmpl, err := template.New("docs-gen").Funcs(funcMap).Parse(docsTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &docsWriter{designPkg: pkg, tmpl: tmpl}, nil
}

func (w *docsWriter) Source() (*GeneratorSource, error) {
	return &GeneratorSource{}, nil
}

const docsTmpl = ``
