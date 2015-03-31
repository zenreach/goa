package writers

import (
	"fmt"
	"text/template"
)

// Handlers writer.
// Helps bootstrap a new app.
type handlersGenWriter struct {
	designPkg     string
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
		designPkg:     designPkg,
		headerGenTmpl: headerT,
		resourceTmpl:  resourceT,
	}, nil
}

func (w *handlersGenWriter) FunctionName() string {
	return ""
}

func (w *handlersGenWriter) Source() string {
	return ""
}

var headerGenTmpl = `
package {{.designPkg}}

import (
	"github.com/raphael/goa"
)

`

var resourceTmpl = ``
