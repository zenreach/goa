package writers

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/raphael/goa/design"

	"gopkg.in/alecthomas/kingpin.v1"
)

// Handlers writer.
// Helps bootstrap a new app.
type handlersGenWriter struct {
	DesignPkg      string
	handlerGenTmpl *template.Template
	interfaceTmpl  string
	dataTypesTmpl  string
}

// NewHandlerWriter returns a writer that produces skeleton code
func NewHandlersGenWriter(designPkg string) (Writer, error) {
	funcMap := template.FuncMap{
		"comment":     comment,
		"commandLine": commandLine,
	}
	headerT, err := template.New("handlers").Funcs(funcMap).Parse(handlerGenTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create handlers template, %s", err)
	}
	return &handlersGenWriter{
		DesignPkg:      designPkg,
		handlerGenTmpl: headerT,
		interfaceTmpl:  interfaceTmpl,
		dataTypesTmpl:  dataTypesTmpl,
	}, nil
}

func (w *handlersGenWriter) FunctionName() string {
	return "genHandlers"
}

func (w *handlersGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.handlerGenTmpl.Execute(&buf, w), "handlers-gen template")
	return buf.String()
}

var handlerGenTmpl = `
var handlerInterfaceTmpl *template.Template
var handlerDataTypesTmpl *template.Template

func {{.FunctionName}}(resource *design.Resource, output string) error {
	var err error
	if handlerInterfaceTmpl == nil {
		handlerInterfaceTmpl, err = template.New("handler-interface").Parse(HandlerInterfaceTmpl)
		if err != nil {
			return fmt.Errorf("failed to create handler interface template, %s", err)
		}
	}
	if handlerDataTypesTmpl == nil {
		funcMap := template.FuncMap{"parameters": parameters, "joinNames": joinNames, "literal": literal}
		handlerDataTypesTmpl, err = template.New("handler-data-types").Funcs(funcMap).Parse(HandlerDataTypesTmpl)
		if err != nil {
			return fmt.Errorf("failed to create handler data type template, %s", err)
		}
	}
	err = os.MkdirAll(output, 0755)
	lowerRes := strings.ToLower(resource.Name)
	w, err := os.Create(path.Join(output, "gen_"+lowerRes+"_handler.go"))
	if err != nil {
		return fmt.Errorf("failed to create output file: %s", err)
	}
	if err := handlerInterfaceTmpl.Execute(w, resource); err != nil {
		return fmt.Errorf("failed to generate %s handler interface: %s", resource.Name, err)
	}
	if err := handlerDataTypesTmpl.Execute(w, resource); err != nil {
		return fmt.Errorf("failed to generate %s handler data types: %s", resource.Name, err)
	}
	return nil
}

const HandlerInterfaceTmpl = ` + "`" + `
{{.RouterTmpl}}
` + "`" + `

const HandlerDataTypesTmpl = ` + "`" + `
{{.MiddlewareTmpl}}
` + "`" + `
`

// Generate parameters signature for given action
func parameters(a *design.Action) string {

}

var interfaceTmpl = `
// {{.Name}} handler interface
type {{.Name}}Handler interface {{{range .Actions}}
	{{capitalize .Name}}({{parameters .}}) *goa.Response{{end}}
}`

var dataTypesTmpl = ``
