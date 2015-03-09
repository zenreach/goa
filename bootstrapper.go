package goa

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/raphael/goa/design"
)

const (
	// Relative path to generated code
	codegenFileName = "goa_handlers.go"

	// bootstrap flag used by goagen to run code generation
	bootstrapFlag = "--bootstrap="
)

// The bootstrapper produces go code that glues the http router with the application controllers.
// It accesses the controller data structures initialized by the application to emit the code.
// The bootstrapper is invoked by running the application with the special "--bootstrap" flag.
// The value of this flag is the path to the directory where the bootstrapper generates the code.
type bootstrapper struct {
	// Writer for generated code
	codegenFile string
	// Template used to generate the code
	tmpl *template.Template
}

// Create bootstrapper.
func newBootstrapper(codegenPath string) (*bootstrapper, error) {
	codegenFile := path.Join(codegenPath, codegenFileName)
	// Make sure file can be written to and is empty
	f, err := os.Create(codegenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file, %s", err.Error())
	}
	f.Close()
	funcMap := template.FuncMap{"joinNames": joinNames}
	tmpl, err := template.New("goagen").Funcs(funcMap).Parse(handlerTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err.Error())
	}
	return &bootstrapper{codegenFile: codegenFile, tmpl: tmpl}, nil
}

// Bootstrap checks whether the --bootstrap command line flag is present and if
// so generate the handlers code and recompiles the app.
func (b *bootstrapper) process(c *controller) error {
	f, err := os.OpenFile(b.codegenFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open output file, %s", err.Error())
	}
	r := c.resource
	for _, a := range r.Actions {
		data := actionData{controller: c, action: a}
		err = b.tmpl.Execute(f, &data)
		if err != nil {
			return fmt.Errorf("failed to generate code, %s", err.Error())
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
	controller *controller
	action     *design.Action
}

func (d *actionData) FuncName() string {
	return d.action.Name + d.controller.resource.Name
}

func (d *actionData) ControllerStruct() string {
	name := reflect.TypeOf(d.controller).String()
	if elems := strings.Split(name, "."); len(elems) > 1 {
		name = elems[1]
	}
	return name
}

func (d *actionData) PathParams() design.ActionParams {
	return d.action.PathParams
}

func (d *actionData) QueryParams() design.ActionParams {
	return d.action.QueryParams
}

func (d *actionData) Payload() *design.Blueprint {
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

const handlerTmpl = `
func {{.FuncName}}(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	c := &{{.ControllerStruct}}{w: w, r: r}
	{{if .PathParams}}// Load and validate path parameters
	{{range $name, $param := .PathParams}}{{$name}}, err := {{$param.TypeName}}.Load(params.ByName("{{$name}}"))
		if err != nil {
			c.RespondBadRequest(err.Error())
		}
	{{end}}{{end}}
	{{if .QueryParams}}// Load and validate query parameters
	{{end}}
	{{if .Payload}}// Load and validate payload
	{{end}}
	// Call controller {{.FuncName}} method
	r := c.{{.FuncName}}({{joinNames .PathParams}}{{if .PathParams}}, {{end}}{{joinNames .QueryParams}}{{if .QueryParams}}{{end}}{{if .Payload}}, payload{{end}})
	{{if .Responses}}var ok = false
	{{range .Responses}}if r.Status == {{.Status}} {
		ok = true{{if .Headers}}
		// VALIDATE HEADERS
	{{end}}{{end}} }
	if !ok {
		c.RespondInternalError(fmt.Printf("API bug, code produced unknown status code %d", r.Status))
		return
	}
	{{end}}
	var b []byte
	if len(r.Body) > 0 {
		var err error
		if b, err = json.Marshal(r.Body); err != nil {
			c.RespondInternalError(fmt.Errorf("API bug, failed to serialize response body: %s", err.Error()))
			return
		} 
	}
	if len(r.Headers) > 0 {
		h := w.Header()
		for n, v := range r.Headers {
			h.Set(n, v)
		}
	}
	w.WriteHeader(r.Status)
	w.Write(b)
}
`
