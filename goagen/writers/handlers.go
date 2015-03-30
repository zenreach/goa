package writers

import (
	"fmt"
	"text/template"
)

// Handlers writer.
// Helps bootstrap a new app.
type handlersWriter struct {
	designPkg    string
	headerTmpl   *template.Template
	resourceTmpl *template.Template
}

// NewHandlerWriter returns a writer that produces skeleton code
func NewHandlersWriter(designPkg string) (Writer, error) {
	funcMap := template.FuncMap{
		"comment":     comment,
		"commandLine": commandLine,
	}
	t := header(fmt.Sprintf("%s handlers", designPkg)) + headerTmpl
	headerT, err := template.New("handlers").Funcs(funcMap).Parse(t)
	resourceT, err := template.New("handlers-resource").Funcs(funcMap).Parse(resourceTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &handlersWriter{
		designPkg:    designPkg,
		headerTmpl:   headerT,
		resourceTmpl: resourceT,
	}, nil
}

func (w *handlersWriter) Source() (*GeneratorSource, error) {
	return &GeneratorSource{}, nil
}

var headerTmpl = `
package {{.designPkg}}

import (
	"github.com/raphael/goa"
)

`

var resourceTmpl = ``
