package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/raphael/goa/goagen/writers"
	"gopkg.in/alecthomas/kingpin.v1"
)

//goagen [--target=gin|negroni|martini|goji] [--bootstrap] [--middleware] [--docs] [--cli=NAME] [--gui]
var (
	designPkg  = kingpin.Flag("pkg", "Package containing Main().").Required()
	outDir     = kingpin.Flag("output", "Path to output directory.").Short('o')
	target     = kingpin.Flag("target", "Web framework code generation should target.").Enum("gin", "goji", "martini", "negroni")
	middleware = kingpin.Flag("middleware", "Generate application middleware.").Short('m').Bool()
	bootstrap  = kingpin.Flag("bootstrap", "Generate new application skeleton.").Bool()
	docs       = kingpin.Flag("docs", "Generate RAML representation of API.").Bool()
	cli        = kingpin.Flag("cli", "Generate API command line client using given name.").String()
	debug      = kingpin.Flag("debug", "Enable debug mode.").Bool()
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		kingpin.Fatalf("can't retrieve current directory: %s", err)
	}
	outputDir := outDir.Default(cwd).String()
	kingpin.Version("0.0.1")
	kingpin.FatalIfError(kingpin.Parse())
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		kingpin.Fatalf("can't create dir %s: %s", *outputDir, err)
	}
	var writers []writers.Writer
	if *middleware {
		writers = append(writers, writers.NewMiddlewareWriter(*designPkg, *target))
	}
	if *bootstrap {
		writers = append(writers, writers.NewBootstrapWriter(*designPkg, *target))
	}
	if *docs {
		writers = append(writers, writers.NewDocsWriter(*designPkg))
	}
	if *cli != "" {
		writers = append(writers, writers.NewCliWriter(*designPkg, *cli))
	}
	for _, w := range writers {
		kingpin.FatalIfError(gen(w))
	}
}

// Generate code using given writer
func gen(writer writers.Writer) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)
	fmt.Printf("-- %s --\n", w.Title())
	report, err := w.Write(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to generate code: %s", e)
	}
	for _, gen := range report.Generated {
		fmt.Printf("%s\n", gen)
	}
	for _, warn := range report.Warnings {
		fmt.Printf("!! %s\n", warn)
	}
	if err := run(tmpDir, *debug, "go", "build", "-o", "goagen"); err != nil {
		return err
	}
	if err := run(tmpDir, *debug, "goagen", "-o", *outputDir); err != nil {
		return fmt.Errof("failed to run code generator: %s", err)
	}
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
}
