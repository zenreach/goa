package writers

import (
	"fmt"
	"text/template"
)

// Doc writer.
type cliGenWriter struct {
	designPkg string
	tmpl      *template.Template
}

// Create middleware writer.
func NewCliGenWriter(pkg string) (Writer, error) {
	tmpl, err := template.New("cli-gen").Parse(cliGenTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &cliGenWriter{designPkg: pkg, tmpl: tmpl}, nil
}

func (w *cliGenWriter) FunctionName() string {
	return ""
}

func (w *cliGenWriter) Source() string {
	return ""
}

const cliGenTmpl = ``
