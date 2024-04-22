//go:build rewrites
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"example.com/example/changing_inputs"
	"example.com/example/directives"
	"example.com/example/flowcontrol"
	"example.com/example/inputs"
	"example.com/example/k8s"
	"example.com/example/labels"
	"example.com/example/mutability"
	"example.com/example/sprig"
	"example.com/example/syntax"
	"example.com/example/typing"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func main() {
	enc := json.NewEncoder(os.Stdout)
	dec := json.NewDecoder(os.Stdin)

	for {
		var dot helmette.Dot
		err_1 := dec.Decode(&dot)
		if err_1 != nil {
			if err_1 == io.EOF {
				break
			}
			panic(err_1)
		}
		tmp_tuple_1 := helmette.Compact2(runChart(&dot))
		err := tmp_tuple_1.T2
		out := tmp_tuple_1.T1

		if out == nil {
			out = map[string]any{}
		}

		if err != nil {
			err = fmt.Sprintf("%+v", err)
		}
		err_2 := enc.Encode(map[string]any{
			"result": out,
			"err":    err,
		})
		if err_2 !=
			nil {
			panic(err_2)
		}
	}
}

func runChart(dot *helmette.Dot) (_ map[string]any, err any) {
	defer func() { err = recover() }()

	switch dot.Chart.Name {
	case "labels":
		return map[string]any{
			"FullLabels": labels.FullLabels(dot),
		}, nil
	case "sprig":
		return map[string]any{
			"Sprig": sprig.Sprig(),
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
			"K8s": k8s.K8s(),
		}, nil

	case "flowcontrol":
		return map[string]any{
			"FlowControl": flowcontrol.FlowControl(dot),
		}, nil

	case "inputs":
		return map[string]any{
			"Inputs": inputs.Inputs(dot),
		}, nil

	case "changing_inputs":
		return map[string]any{
			"ChangingInputs": changing_inputs.ChangingInputs(dot),
		}, nil

	case "syntax":
		return map[string]any{
			"Syntax": syntax.Syntax(),
		}, nil

	default:
		panic(fmt.Sprintf("unknown package %q", dot.Chart.Name))
	}
}
