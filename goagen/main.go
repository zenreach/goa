package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/raphael/goa/goagen/writers"
	"gopkg.in/alecthomas/kingpin.v1"
)

// goagen [--pkg=PKG] [--input|-i=INPUT] [--output|-o=OUTPUT] [--handlers|-h] [--middleware|-m]
//        [--docs|-d] [--cli|-c=CLI] [--gui|-g] [--debug]
var (
	designPkg  = kingpin.Flag("pkg", "Design package containing Init().").Default("resources").Short('p').String()
	inDir      = kingpin.Flag("input", "Path to directory containing application design package source.").Short('i')
	outDir     = kingpin.Flag("output", "Path to output directory.").Short('o')
	handlers   = kingpin.Flag("handlers", "Generate new application skeleton.").Short('h').Bool()
	middleware = kingpin.Flag("middleware", "Generate application middleware.").Short('m').Bool()
	docs       = kingpin.Flag("docs", "Generate RAML representation of API.").Short('d').Bool()
	cli        = kingpin.Flag("cli", "Generate API command line client using given name.").Short('c').String()
	debug      = kingpin.Flag("debug", "Enable debug mode.").Bool()
	nobuild    = kingpin.Flag("nobuild", "Write generator code only, do not run it.").Bool()
	inputDir   string
	outputDir  string
	buildDir   string
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		kingpin.Fatalf("can't retrieve current directory: %s", err)
	}
	inputDirOpt := inDir.Default(cwd).String()
	outputDirOpt := outDir.Default(cwd).String()
	kingpin.Version("0.0.1")
	kingpin.Parse()
	inputDir = *inputDirOpt
	outputDir = *outputDirOpt
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		kingpin.Fatalf("can't create dir %s: %s", outputDir, err)
	}
	bDir, err := ioutil.TempDir("", "")
	if err != nil {
		kingpin.Fatalf("failed to create temp dir: %s", err)
	}
	buildDir = bDir
	err = setupFiles()
	if err == nil {
		err = writeGenerator()
	}
	if err == nil && !*nobuild {
		err = runGenerator()
	}
	if err == nil && *nobuild {
		fmt.Printf("Generator written at %s\n", buildDir)
	} else {
		os.RemoveAll(buildDir)
	}
	kingpin.FatalIfError(err, "")
}

// setupFiles copies all application design files to the build directory
func setupFiles() error {
	designDir := path.Join(buildDir, *designPkg)
	os.MkdirAll(designDir, 0755)
	return moveFiles(inputDir, designDir)
}

// Helper function that moves all files from one directory to another recursively
func moveFiles(from, to string) error {
	return filepath.Walk(from, func(abs string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("can't load design package files: %s", err)
		}
		rel := strings.TrimPrefix(abs, from)
		dest := path.Join(to, rel)
		if info.IsDir() {
			os.MkdirAll(dest, 0755)
		} else {
			b, err := ioutil.ReadFile(abs)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(dest, b, 0755); err != nil {
				return err
			}
		}
		return nil
	})
}

// writeGenerator writes the go source for the generator into buildDir.
func writeGenerator() error {
	var ws []writers.Writer
	all := !*middleware && !*handlers && !*docs && *cli == ""
	if *middleware || all {
		if w, err := writers.NewMiddlewareGenWriter(); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	if *handlers || all {
		if w, err := writers.NewHandlersGenWriter(*designPkg); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	if *docs || all {
		if w, err := writers.NewDocsGenWriter(*designPkg); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	if *cli != "" || all {
		if w, err := writers.NewCliGenWriter(*designPkg); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	goagenT, err := template.New("goagen").Parse(goagenTmpl)
	if err != nil {
		return fmt.Errorf("failed to create goagen template: %s", err)
	}
	mainPath := path.Join(buildDir, "main.go")
	f, err := os.Create(mainPath)
	if err != nil {
		return fmt.Errorf("Cannot create %s: %s", mainPath, err)
	}
	defer f.Close()
	d := genData{Package: *designPkg, Writers: ws}
	if err := goagenT.Execute(f, d); err != nil {
		return fmt.Errorf("Failed to render generator code: %s", err)
	}
	return nil
}

// runGenerator compiles and runs the code generator
func runGenerator() error {
	env := os.Environ()
	env = append(env, "GOPATH="+os.Getenv("GOPATH")+":"+buildDir)
	if out, err := run(buildDir, *debug, env, "go", "build", "-o", "goagen"); err != nil {
		return fmt.Errorf("%s\n%s", err, out)
	}
	tmpOut, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpOut)
	if out, err := run(buildDir, *debug, env, "./goagen", "-o", tmpOut); err != nil {
		return fmt.Errorf("%s\n%s", err, out)
	}
	cmd := exec.Command("go", "fmt")
	cmd.Dir = tmpOut
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(b))
	}
	return moveFiles(tmpOut, outputDir)
}

// run runs given command at given location optionally printing the command to the console.
func run(dir string, debug bool, env []string, path string, args ...string) (string, error) {
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}
	if debug {
		cmdStr := fmt.Sprintf("%s> %s %s", dir, path, strings.Join(args, " "))
		fmt.Printf("%s\n", cmdStr)
	}
	output, err := cmd.CombinedOutput()
	if debug {
		fmt.Printf("%s\n", output)
	}
	if err != nil {
		return string(output), fmt.Errorf("%s exited with %s", path, err)
	}
	return string(output), nil
}

// Data used to render template
type genData struct {
	Package string
	Writers []writers.Writer
}

const goagenTmpl = `
package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"./{{.Package}}"
	"github.com/raphael/goa/design"
)
{{$pkg := .Package}}
func main() {
	{{$pkg}}.Init()
	if len(os.Args) < 3 || os.Args[1] != "-o" {
		fmt.Fprintf(os.Stderr, "usage: %s -o OUTPUT_DIR", os.Args[0])
		os.Exit(1)
	}
	output := os.Args[2]
	for _, n := range design.ResourceNames {
		res := design.Resources[n]
		err := res.Validate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", n, err)
			os.Exit(1)
		}{{range .Writers}}
		err = {{.FunctionName}}(res, output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", n, err)
			os.Exit(1)
		} {{end}}
	}
}

func joinNames(params design.ActionParams) string {
	names := make([]string, len(params))
	idx := 0
	for n, _ := range params {
		names[idx] = n
		idx += 1
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func literal(val interface{}) string {
	return fmt.Sprintf("%#v", val)
}

{{range .Writers}}
{{.Source}}
{{end}}
`
