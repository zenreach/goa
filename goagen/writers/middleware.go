package writers

import (
	"bytes"
	"fmt"
	"text/template"

	"gopkg.in/alecthomas/kingpin.v1"
)

// Middleware writer.
type middlewareGenWriter struct {
	genTmpl        *template.Template
	MiddlewareTmpl string
	RouterTmpl     string
}

// Create middleware writer.
func NewMiddlewareGenWriter() (Writer, error) {
	genTmpl, err := template.New("middleware-gen").Parse(middlewareGenTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware-gen template, %s", err)
	}
	return &middlewareGenWriter{
		genTmpl: genTmpl,
		MiddlewareTmpl: middlewareTmpl,
		RouterTmpl: routerTmpl,
	}, nil
}

func (w *middlewareGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.genTmpl.Execute(&buf, w), "middleware-gen template")
	return buf.String()
}

func (w *middlewareGenWriter) FunctionName() string {
	return "genMiddleware"
}

const middlewareGenTmpl = `
var resRouterTmpl *template.Template
var resMiddlewareTmpl *template.Template

func {{.FunctionName}}(resource *design.Resource, output string) error {
	var err error
	if resRouterTmpl == nil {
		resRouterTmpl, err = template.New("router").Parse(RouterTmpl)
		if err != nil {
			return fmt.Errorf("failed to create router template, %s", err)
		}
	}
	if resMiddlewareTmpl == nil {
		funcMap := template.FuncMap{"parameters": parameters, "joinNames": joinNames, "literal": literal}
		resMiddlewareTmpl, err = template.New("middleware").Funcs(funcMap).Parse(MiddlewareTmpl)
		if err != nil {
			return fmt.Errorf("failed to create middleware template, %s", err)
		}
	}
	err = os.MkdirAll(output, 0755)
	lowerRes := strings.ToLower(resource.Name)
	w, err := os.Create(path.Join(output, "gen_"+lowerRes+"_middleware.go"))
	if err != nil {
		return fmt.Errorf("failed to create output file: %s", err)
	}
	if err := resRouterTmpl.Execute(w, resource); err != nil {
		return fmt.Errorf("failed to generate %s router: %s", resource.Name, err)
	}
	if err := resMiddlewareTmpl.Execute(w, resource); err != nil {
		return fmt.Errorf("failed to generate %s middleware: %s", resource.Name, err)
	}
	return nil
}

// Helper function that generates an action call site parameters.
func parameters(a *design.Action) string {
	var params []string
	if a.Payload != nil {
		params = append(params, "&payload")
	}
	pathParams := make([]string, len(a.PathParams))
	i := 0
	for n, _ := range a.PathParams {
		pathParams[i] = n
		i += 1
	}
	sort.Strings(pathParams)
	params = append(params, pathParams...)
	queryParams := make([]string, len(a.QueryParams))
	i = 0
	for n, _ := range a.QueryParams {
		queryParams[i] = n
		i += 1
	}
	sort.Strings(queryParams)
	params = append(params, queryParams...)
	return strings.Join(params, ", ")
}

const RouterTmpl = ` + "`" + `
{{.RouterTmpl}}
` + "`" + `

const MiddlewareTmpl = ` + "`" + `
{{.MiddlewareTmpl}}
` + "`" + `
`
const routerTmpl = `
package main

import (
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa"
)

func {{.Name}}Router() { {{$resource := .}}
	router := httpRouter.New(){{range $actionName, $action := .Actions}}
	router.{{$action.HttpMethod}}(path.Join("{{$resource.BasePath}}", "{{$action.Path}}"), {{$action.Name}}{{$resource.Name}}){{end}}
	return router
}`

const middlewareTmpl = `{{$resource := .}}{{range $actionName, $action := .Actions}}
func {{$actionName}}{{$resource.Name}}(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := goa.New{{$resource.Name}}Handler(w, r){{range $name, $param := $action.PathParams}}
	{{$name}}, err := {{$param.Member.Type.Name}}.Load(params.ByName("{{$name}}"))
	if err != nil {
		goa.RespondBadRequest(w, "Invalid param '{{$name}}': %s", err)
		return
	}{{end}}{{/* range $action.PathParams */}}{{if $action.QueryParams}}
	query := r.URL.Query()
	{{range $name, $param := $action.QueryParams}}{{$name}}, err := {{$param.Member.Type.Name}}.Load(query["{{$name}}"]{{if not (eq $param.Member.Type.Name "array")}}[0]{{end}})
	if err != nil {
		goa.RespondBadRequest(w, "Invalid param '{{$name}}': %s", err)
		return
	}
	{{end}}{{end}}{{/* if $action.QueryParams */}}{{if $action.Payload}}
	b, err := h.LoadRequestBody(r)
	if err != nil {
		goa.RespondBadRequest(w, err)
		return
	}
	raw, err := res.Actions["{{$actionName}}"].Payload.Load("payload", b)
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	var payload {{$actionName}}Payload
	err = goa.InitStruct(&payload, raw.(map[string]interface{}))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}{{end}}{{/* if $action.Payload */}}
	resp := h.{{$actionName}}({{parameters $action}})
	if resp == nil {
		// Response already written by handler
		return
	}
	{{if .Responses}}ok := resp.Status == 400 || resp.Status == 500
	if !ok {
		{{range $action.Responses}}if resp.Status == {{.Status}} {
			ok = true{{if .MediaType}}
			resp.Header.Set("Content-Type", "{{.MediaType.Identifier}}+json"){{end}}{{if .HeaderPatterns}}
			var h string
			{{range $name, $value := .HeaderPatterns}}h = resp.Header.Get("{{$name}}")
			if !regexp.MatchString("{{$value}}", h) {
				goa.RespondInternalError(w, fmt.Printf("API bug, code produced invalid {{$name}} header value.", h))
				return
			}{{end}}{{end}}
		}
	{{end}}}
	if !ok {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	{{end}}{{/* if .Responses */}}
	resp.Write(w)
}
{{end}}
`
