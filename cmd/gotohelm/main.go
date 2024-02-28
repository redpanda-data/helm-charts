package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm"
	"golang.org/x/tools/go/packages"
)

func main() {
	out := flag.String("write", "-", "The directory to write the transpiled templates to or - to write them to standard out")

	flag.Parse()

	cwd, _ := os.Getwd()

	pkgs, err := gotohelm.LoadPackages(&packages.Config{
		Dir: cwd,
	}, flag.Args()...)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		chart, err := gotohelm.Transpile(pkg)
		if err != nil {
			fmt.Printf("Failed to transpile %q: %s\n", pkg.Name, err)
			continue
		}

		if *out == "-" {
			writeToStdout(chart)
		} else {
			if err := writeToDir(chart, *out); err != nil {
				panic(err)
			}
		}

	}
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
		file, err := os.OpenFile(path.Join(dir, f.Name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
