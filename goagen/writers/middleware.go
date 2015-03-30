package writers

import (
	"bytes"
	"fmt"
	"text/template"
)

// Middleware writer.
type middlewareWriter struct {
	designPkg string
	tmpl      *template.Template
}

// Create middleware writer.
func NewMiddlewareWriter(pkg string) (Writer, error) {
	tmpl, err := template.New("middleware-gen").Parse(middlewareGenTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware-gen template, %s", err)
	}
	return &middlewareWriter{designPkg: pkg, tmpl: tmpl}, nil
}

func (w *middlewareWriter) Source() (string, error) {
	var buf bytes.Buffer
	err := w.tmpl.Execute(&buf, w)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (w *middlewareWriter) FunctionName() string {
	return fmt.Sprintf("gen%sMiddleware", w.designPkg)
}

const middlewareGenTmpl = `
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

var resTmpl *template.Template

func {{.FunctionName}}(resource *design.Resource) error {
	if resTmpl == nil {
		funcMap := template.FuncMap{"joinNames": joinNames, "literal": literal}
		resTmpl, err := template.New("middleware").Funcs(funcMap).Parse(middlewareTmpl)
		if err != nil {
			return fmt.Errorf("failed to create middleware template, %s", err)
		}
	}
	if err := tmpl.Execute(resource, w); err != nil {
		return fmt.Errorf("failed to generate %s middleware: %s", name, err)
	}
}
`

const middlewareTmpl = `
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
	resp := h.{{.action.FuncName}}({{if .action.Payload}}payload{{end}}{{if .action.PathParams}}, {{joinNames .action.PathParams}}{{end}}{{if .action.QueryParams}}{{joinNames .action.QueryParams}}{{end}})
	if resp == nil {
		// Response already written by handler
		return
	}
	{{if .Responses}}ok := resp.Status == 400 || resp.Status == 500
	if !ok {
		{{range .Responses}}if resp.Status == {{.Status}} {
			ok = true{{if .MediaType}}
			resp.Header.Set("Content-Type", "{{.MediaType.Identifier}}+json"){{$name, $value := range .Headers}}
			{{end}}resp.Header.Set("{{$name}}", "{{$value}}")
		}{{end}}
	}{{$name, $value := range .HeaderPatterns}}
	h := resp.Header.Get("{{$name}}")
	if !regexp.MatchString("{{$value}}", h) {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced invalid ${{name}} header value.", h))
		return
	}{{end}}	
	{{end}} }
	if !ok {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	{{end}}{{/* if .Responses */}}
	goa.WriteResponse(w, r)
}
`
