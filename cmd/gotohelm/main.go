package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm"
	"golang.org/x/tools/go/packages"
)

func main() {
	out := flag.String("write", "-", "The directory to write the transpiled templates to or - to write them to standard out")

	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Printf("usage: gotohelm <package to transpile> [dependencies...]")
		os.Exit(1)
	}

	cwd, _ := os.Getwd()

	pkgs, err := gotohelm.LoadPackages(&packages.Config{
		Dir: cwd,
	}, flag.Args()[0])
	if err != nil {
		panic(err)
	}

	if len(pkgs) != 1 {
		fmt.Printf("loading %q resulted in loading more than one package.", flag.Args()[0])
		os.Exit(1)
	}

	pkg := pkgs[0]

	deps, err := goList(flag.Args()[1:]...)
	if err != nil {
		panic(err)
	}

	chart, err := gotohelm.Transpile(pkg, deps...)
	if err != nil {
		fmt.Printf("Failed to transpile %q: %s\n", pkg.Name, err)
		os.Exit(1)
	}

	if *out == "-" {
		writeToStdout(chart)
	} else {
		if err := writeToDir(chart, *out); err != nil {
			panic(err)
		}
	}
}

func goList(patterns ...string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	cmd := exec.Command("go", append([]string{"list"}, patterns...)...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(out), "\n"), nil
}

func writeToStdout(chart *gotohelm.Chart) {
	for _, f := range chart.Files {
		fmt.Printf("%s\n", f.Name)
		f.Write(os.Stdout)
		fmt.Printf("\n\n")
	}
}

func writeToDir(chart *gotohelm.Chart, dir string) error {
	for _, f := range chart.Files {
		file, err := os.OpenFile(path.Join(dir, f.Name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}

		f.Write(file)

		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}
