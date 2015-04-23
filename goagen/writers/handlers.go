package writers

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/raphael/goa/design"

	"gopkg.in/alecthomas/kingpin.v1"
)

// Handlers writer.
// Helps bootstrap a new app.
type handlersGenWriter struct {
	DesignPkg     string
	genTmpl       *template.Template
	InterfaceTmpl string
	DataTypesTmpl string
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
		DesignPkg:     designPkg,
		genTmpl:       headerT,
		InterfaceTmpl: interfaceTmpl,
		DataTypesTmpl: dataTypesTmpl,
	}, nil
}

func (w *handlersGenWriter) FunctionName() string {
	return "genHandlers"
}

func (w *handlersGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.genTmpl.Execute(&buf, w), "handlers-gen template")
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
{{.InterfaceTmpl}}
` + "`" + `

const HandlerDataTypesTmpl = ` + "`" + `
{{.DataTypesTmpl}}
` + "`" + `
`

var interfaceTmpl = `
// {{.Name}} handler interface{{$resource := .}}
type {{.Name}}Handler interface { {{range .Actions}}
	{{capitalize .Name}}({{parameters $resource .}}) *goa.Response{{end}}
}`

var dataTypesTmpl = `{{range .Actions}}{{$type := payloadType .Payload}}{{if $type}}


type {{.Name}}Payload struct {
}`

// Compute payload type of given payload.
// Return array element type if payload is an array,
func payloadType(payload *design.Member) design.Object {
	elemType := payload.Type
	for elemType.Kind() == design.ArrayType {
		elemType = elemType.(*design.Array).ElemType
	}
	o, ok := elemType.(design.Object)
	if ok {
		return o
	}
	return nil
}

// Go parameters for action method
func parameters(r *design.Resource, a *design.Action) string {
	var params []string
	if a.Payload != nil {
		params = append(params, fmt.Sprintf("payload %s", signature(r, a, "Payload", a.Payload.Type)))
	}
	for _, n := range a.PathParamNames() {
		p := a.PathParams[n]
		params = append(params, fmt.Sprintf("%s %s", n, signature(r, a, p.Name, p.Member.Type)))
	}
	for _, n := range a.QueryParamNames() {
		p := a.PathParams[n]
		params = append(params, fmt.Sprintf("%s %s", n, signature(r, a, p.Name, p.Member.Type)))
	}
	return strings.Join(params, ", ")
}

func signature(r *design.Resource, a *design.Action, suffix string, t design.DataType) string {
	switch t.Kind() {
	case design.BooleanType:
		return "bool"
	case design.IntegerType:
		return "int"
	case design.NumberType:
		return "float64"
	case design.StringType:
		return "string"
	case design.ArrayType:
		ar := t.(*design.Array)
		return "[]" + signature(r, a, suffix, ar.ElemType)
	case design.ObjectType:
		return fmt.Sprintf("*%s%s%s", r.Name, a.Name, suffix)
	}
	kingpin.Fatalf("Unknown or invalid type '%v'", t.Kind())
	return "" // Not reached
}
