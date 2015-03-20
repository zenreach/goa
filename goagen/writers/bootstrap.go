package writers

import (
	"fmt"
	"text/template"
)

type bootstrapWriter struct {
	designPkg    string
	target       string
	headerTmpl   *template.Template
	resourceTmpl *template.Template
}

// NewBootstrapWriter returns a writer that produces skeleton code for the given target web
// framework. The currently supported frameworks are goa (default), gin, goji and martini.
func NewBootstrapWriter(designPkg, target string) (Writer, error) {
	funcMap := template.FuncMap{
		"comment":     comment,
		"commandLine": commandLine,
	}
	var tmpl *template.Template
	var err error
	switch target {
	case "goa":
		t := header(fmt.Sprintf("%s Goa handlers", designPkg)) + goaHeaderTmpl
		headerTmpl, err = template.New("goa-bootstrap").Funcs(funcMap).Parse(t)
		resourceTmpl, err = template.New("goa-bootstrap-resource").Funcs(funcMap).Parse(goaResourceTmpl)
	case "gin":
		t := header(fmt.Sprintf("%s Gin handlers", designPkg)) + ginHeaderTmpl
		headerTmpl, err = template.New("gin-bootstrap").Funcs(funcMap).Parse(t)
		resourceTmpl, err = template.New("gin-bootstrap-resource").Funcs(funcMap).Parse(ginResourceTmpl)
	case "goji":
		t := header(fmt.Sprintf("%s Goji handlers", designPkg)) + gojiHeaderTmpl
		headerTmpl, err = template.New("goji-bootstrap").Funcs(funcMap).Parse(t)
		resourceTmpl, err = template.New("goji-bootstrap-resource").Funcs(funcMap).Parse(gojiResourceTmpl)
	case "martini":
		t := header(fmt.Sprintf("%s Martini handlers", designPkg)) + martiniHeaderTmpl
		headerTmpl, err = template.New("martini-bootstrap").Funcs(funcMap).Parse(t)
		resourceTmpl, err = template.New("martini-bootstrap-resource").Funcs(funcMap).Parse(martiniResourceTmpl)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create template, %s", err)
	}
	return &bootstrapWriter{designPkg: designPkg, target: target, tmpl: tmpl}, nil
}

var goaHeaderTmpl = `
package {{.designPkg}}

import (
	"github.com/raphael/goa"
)

`

var goaResourceTmpl = `

var ginHeaderTmpl = `
package {{.designPkg}}

import (
	"github.com/gin-gonic/gin"
)

`

var gojiHeaderTmpl = `
package {{.designPkg}}

import (
	"github.com/zenazn/goji"
)

`

var martiniHeaderTmpl = `
package {{.designPkg}}

import (
	"github.com/go-martini/martini"
)

`
