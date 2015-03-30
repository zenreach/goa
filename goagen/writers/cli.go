package writers

import (
	"fmt"
	"text/template"
)

// Doc writer.
type cliWriter struct {
	designPkg string
	tmpl      *template.Template
}

// Create middleware writer.
func NewCliWriter(pkg string) (Writer, error) {
	funcMap := template.FuncMap{"joinNames": joinNames, "literal": literal}
	tmpl, err := template.New("cli-gen").Funcs(funcMap).Parse(cliTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &cliWriter{designPkg: pkg, tmpl: tmpl}, nil
}

func (w *cliWriter) Source() (*GeneratorSource, error) {
	return &GeneratorSource{}, nil
}

const cliTmpl = ``
