package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func main() {
	srcDir, err := os.Getwd()
	if err != nil {
		fail(err.Error())
	}
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		fail(err.Error())
	}
	goagen := path.Join(tmpDir, "goagen")
	debug := len(os.Args) > 1 && os.Args[1] == "--debug"
	run(exec.Command("go", "build", "-o", goagen), debug)
	run(exec.Command(goagen, fmt.Sprintf("--bootstrap=%s", srcDir)), debug)
}

// Helper method that runs command and calls fail if command fails
func run(cmd *exec.Cmd, debug bool) {
	if debug {
		fmt.Printf("%s\n", strings.Join(cmd.Args, " "))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fail(string(output) + "\n" + err.Error())
	}
}

// Fail with given error message
func fail(msg string) {
	fmt.Printf(msg)
	os.Exit(1)
}
