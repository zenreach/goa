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
	designPkg  = kingpin.Flag("pkg", "Package containing Init().").Default("design").String()
	inDir      = kingpin.Flag("input", "Path to directory containing application design package source.").Short('i')
	outDir     = kingpin.Flag("output", "Path to output directory.").Short('o')
	handlers   = kingpin.Flag("handlers", "Generate new application skeleton.").Short('h').Bool()
	middleware = kingpin.Flag("middleware", "Generate application middleware.").Short('m').Bool()
	docs       = kingpin.Flag("docs", "Generate RAML representation of API.").Short('d').Bool()
	cli        = kingpin.Flag("cli", "Generate API command line client using given name.").Short('c').String()
	debug      = kingpin.Flag("debug", "Enable debug mode.").Bool()
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
	defer os.RemoveAll(buildDir)
	err = setupFiles()
	if err == nil {
		err = writeGenerator()
	}
	if err == nil {
		err = runGenerator()
	}
	kingpin.FatalIfError(err, "")
}

// setupFiles copies all application design files to the build directory
func setupFiles() error {
	err := filepath.Walk(inputDir, func(abs string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("can't load design package files: %s", err)
		}
		rel := strings.TrimPrefix(inputDir, abs)
		dest := path.Join(buildDir, rel)
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
	return err
}

// writeGenerator writes the go source for the generator into buildDir.
func writeGenerator() error {
	var ws []writers.Writer
	all := !*middleware && !*handlers && !*docs && *cli != ""
	if *middleware || all {
		if w, err := writers.NewMiddlewareWriter(*designPkg); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	if *handlers || all {
		if w, err := writers.NewHandlersWriter(*designPkg); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	if *docs || all {
		if w, err := writers.NewDocsWriter(*designPkg); err != nil {
			return err
		} else {
			ws = append(ws, w)
		}
	}
	if *cli != "" || all {
		if w, err := writers.NewCliWriter(*designPkg); err != nil {
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
	f, err := os.Open(mainPath)
	if err != nil {
		return fmt.Errorf("Cannot create %s: %s", mainPath, err)
	}
	defer f.Close()
	d := genData{pkg: *designPkg, writers: ws}
	if err := goagenT.Execute(f, d); err != nil {
		return fmt.Errorf("Failed to render generator code: %s", err)
	}
	return nil
}

// runGenerator compiles and runs the code generator
func runGenerator() error {
	if err := run(buildDir, *debug, "go", "build", "-o", "goagen"); err != nil {
		return fmt.Errorf("failed to build code generator: %s", err)
	}
	if err := run(buildDir, *debug, "goagen", "-o", outputDir); err != nil {
		return fmt.Errorf("failed to run code generator: %s", err)
	}
	return nil
}

// run runs given command at given location optionally printing the command to the console.
func run(dir string, debug bool, path string, args ...string) error {
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	if debug {
		cmdStr := fmt.Sprintf("%s> %s %s", dir, path, strings.Join(args, " "))
		fmt.Printf("%s\n", cmdStr)
	}
	output, err := cmd.CombinedOutput()
	if debug {
		fmt.Printf("%s\n", output)
	}
	if err != nil {
		return fmt.Errorf("%s exited with %s", path, err)
	}
	return nil
}

// Data used to render template
type genData struct {
	pkg     string
	writers []writers.Writer
}

const goagenTmpl = `
import (
	"os"
)
 {{$pkg := .pkg}}
func main() {
	for _, n := range {{$pkg}}.ResourNames {
		res := {{$pkg}}.Resources[n]{{range .writers}}
		err := {{.FunctionName}}(res)
		if err != nil {
			fmt.Sprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		} {{end}}
	}
}
{{range .writers}}
{{.Source}}
{{end}}
`
