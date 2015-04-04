package writers

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alecthomas/kingpin"
)

// Handlers writer.
// Helps bootstrap a new app.
type handlersGenWriter struct {
	DesignPkg     string
	headerGenTmpl *template.Template
	resourceTmpl  *template.Template
}

// NewHandlerWriter returns a writer that produces skeleton code
func NewHandlersGenWriter(designPkg string) (Writer, error) {
	funcMap := template.FuncMap{
		"comment":     comment,
		"commandLine": commandLine,
	}
	t := header(fmt.Sprintf("%s handlers", designPkg)) + headerGenTmpl
	headerT, err := template.New("handlers").Funcs(funcMap).Parse(t)
	resourceT, err := template.New("handlers-resource").Funcs(funcMap).Parse(resourceTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &handlersGenWriter{
		DesignPkg:     designPkg,
		headerGenTmpl: headerT,
		resourceTmpl:  resourceT,
	}, nil
}

func (w *handlersGenWriter) FunctionName() string {
	return "genHandlers"
}

func (w *handlersGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.headerGenTmpl.Execute(&buf, w), "handlers-gen template")
	return buf.String()
}

var headerGenTmpl = `
func {{.FunctionName}}(resource *design.Resource, output string) error {
	return nil
}`

var resourceTmpl = ``
