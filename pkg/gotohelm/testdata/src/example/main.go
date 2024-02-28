package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"example.com/example/a"
	"example.com/example/b"
	"example.com/example/directives"
	"example.com/example/flowcontrol"
	"example.com/example/inputs"
	"example.com/example/k8s"
	"example.com/example/mutability"
	"example.com/example/sprig"
	"example.com/example/typing"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func main() {
	enc := json.NewEncoder(os.Stdout)
	dec := json.NewDecoder(os.Stdin)

	for {
		var dot helmette.Dot
		if err := dec.Decode(&dot); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		out, err := runChart(&dot)

		if out == nil {
			out = map[string]any{}
		}

		if err := enc.Encode(map[string]any{
			"result": out,
			"err":    err,
		}); err != nil {
			panic(err)
		}
	}
}

func runChart(dot *helmette.Dot) (_ map[string]any, err any) {
	defer func() { err = recover() }()

	switch dot.Chart.Name {
	case "sprig":
		return map[string]any{
			"Sprig": sprig.Sprig(),
		}, nil

	case "a":
		return map[string]any{
			"ConfigMap": a.ConfigMap(),
		}, nil

	case "b":
		return map[string]any{
			"Constant":  b.Constant(),
			"ConfigMap": b.ConfigMap(),
		}, nil

	case "typing":
		return map[string]any{
			"Typing": typing.Typing(dot),
		}, nil

	case "directives":
		return map[string]any{
			"Directives": directives.Directives(),
		}, nil

	case "mutability":
		return map[string]any{
			"Mutability": mutability.Mutability(),
		}, nil

	case "k8s":
		return map[string]any{
			"Pod": k8s.Pod(),
			"PDB": k8s.PDB(),
		}, nil

	case "flowcontrol":
		return map[string]any{
			"FlowControl": flowcontrol.FlowControl(dot),
		}, nil

	case "inputs":
		return map[string]any{
			"Inputs": inputs.Inputs(dot),
		}, nil

	default:
		panic(fmt.Sprintf("unknown package %q", dot.Chart.Name))
	}
}
