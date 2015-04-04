package writers

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alecthomas/kingpin"
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
	return "genCli"
}

func (w *cliGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.tmpl.Execute(&buf, w), "cli-gen template")
	return buf.String()
}

const cliGenTmpl = `
func {{.FunctionName}}(resource *design.Resource, output string) error {
	return nil
}`
