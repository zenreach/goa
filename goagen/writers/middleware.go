package writers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/raphael/goa/design"
)

// Create middleware writer.
func NewMiddlewareWriter(pkg, target string)
	funcMap := template.FuncMap{"joinNames": joinNames, "literal": literal}
	tmpl, _ := template.New("goagen").Funcs(funcMap).Parse(handlerTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &bootstrapper{codegenFile: codegenFile, tmpl: tmpl}, nil
}

// Bootstrap checks whether the --bootstrap command line flag is present and if
// so generate the handlers code and recompiles the app.
func (b *bootstrapper) process(c *controller) error {
	f, err := os.OpenFile(b.codegenFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open output file, %s", err)
	}
	r := c.resource
	for _, a := range r.Actions {
		data := actionData{resourceName: r.Name, payloadType: a.Payload.Type.Name, action: a}
		err = b.tmpl.Execute(f, &data)
		if err != nil {
			return fmt.Errorf("failed to generate code, %s", err)
		}
	}
	f.Close()
	o, err := exec.Command("go", "fmt", b.codegenFile).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to format generated code, %s", o)
	}
	return nil
}

// Cleanup any generated file.
func (b *bootstrapper) cleanup() {
	if b.codegenFile != "" {
		os.Remove(b.codegenFile)
	}
}

// Data structure used by template
type actionData struct {
	resourceName string
	payloadType  string // go type of payload
	action       *design.Action
}

func (d *actionData) FuncName() string {
	return d.action.Name + d.controller.resource.Name
}

func (d *actionData) PathParams() design.ActionParams {
	return d.action.PathParams
}

func (d *actionData) QueryParams() design.ActionParams {
	return d.action.QueryParams
}

func (d *actionData) Payload() design.Object {
	return d.action.Payload
}

func joinNames(params design.ActionParams) string {
	var names = make([]string, len(params))
	var idx = 0
	for n, _ := range params {
		names[idx] = n
		idx += 1
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// literal returns the value as if it was declared as a literal value in a go program
func literal(val interface{}) string {
	return fmt.Sprintf("%#v", val)
}

const handlerTmpl = `
func {{.FuncName}}(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := goa.NewHandler("{{.resourceName}}", w, r){{range $name, $param := .action.PathParams}}
	{{$name}}, err := {{$param.TypeName}}.Load(params.ByName("{{$name}}"))
	if err != nil {
		goa.RespondBadRequest(w, "Invalid value for %s: %s", $name, err)
		return
	}{{end}}{{if .QueryParams}}
	query := req.URL.Query()
	{{range $name, $param := .action.QueryParams}}{{$name}}, err := {{$param.TypeName}}.Load(query["{{$name}}"]{{if not (eq $param.Type.Name "array")}}[0]{{end}})
	if err != nil {
		goa.RespondBadRequest(w, "Invalid value for %s: %s", $name, err)
		return
	}
	{{end}}{{end}}{{if .action.Payload}}
	b, err := h.LoadRequestBody(r)
	if err != nil {
		goa.RespondBadRequest(w, err)
		return
	}
	parsed := make(map[string]interface{})
	{{range $name, $prop := .action.Payload}}
	var value interface{}
	raw := values["{{$name}}"]{{if $prop.DefaultValue}}
	if raw == nil {
		raw = {{literal $prop.DefaultValue}}
	}{{end}}
	if raw != nil { {{if not (eq $prop.Type.Name "array")}}
		if reflect.TypeOf(raw).Kind() == reflect.Slice {
			// Extra element from array if necessary (some encodings always produce arrays)
			arr := reflect.ValueOf(raw)
			if arr.Len() > 0 {
				raw = arr.Index(0)
			}
		}{{end}}
		var err error
		value, err = goa.{{$prop.Type.Name}}.Load(raw)
		if err != nil {
			goa.RespondBadRequest(w, "error loading '{{$name}}': %s", err)
			return
		}
		parsed[name] = value
	}
	{{end}}var payload *{{.payloadType}}
	if err := h.InitStruct(payload, parsed); err != nil {
		goa.RespondBadRequest(w, "error initializing payload data structure: %s", err)
		return
	} {{end}}{{/* if .action.Payload */}}
	r := h.{{.action.FuncName}}({{if .action.Payload}}payload{{end}}{{if .action.PathParams}}, {{joinNames .action.PathParams}}{{end}}{{if .action.QueryParams}}{{joinNames .action.QueryParams}}{{end}})
	if r == nil {
		// Response already written by handler
		return
	}
	{{if .Responses}}ok := false
	{{range .Responses}}if r.Status == {{.Status}} {
		ok = true{{if .MediaType}}
		r.Header.Set("Content-Type", "{{.MediaType.Identifier}}+json"){{$name, $value := range .Headers}}
		{{end}}r.Header.Set("{{$name}}", "{{$value}}")
	}{{end}}{{$name, $value := range .HeaderPatterns}}
	h := r.Header.Get("{{$name}}")
	if !regexp.MatchString("{{$value}}", h) {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced invalid ${{name}} header value.", h))
		return
	}{{end}}	
	{{end}} }
	if !ok {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced unknown status code %d", r.Status))
		return
	}
	{{end}}{{/* if .Responses */}}
	goa.WriteResponse(w, r)
}
`
