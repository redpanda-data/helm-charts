package typing

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func typeTesting(dot *helmette.Dot) string {
	t := dot.Values["t"]

	if _, ok := t.(string); ok {
		return "it's a string!"
	} else if _, ok := t.(int); ok {
		return "it's an int!"
	} else if _, ok := t.(float64); ok {
		return "it's a float!"
	}

	return "it's something else!"
}

func typeAssertions(dot *helmette.Dot) string {
	return "Not yet supported"
	// _ = dot.Values["no-such-key"].(int)
	// return "Didn't panic!"
}

func typeSwitching(dot *helmette.Dot) string {
	return "Not yet supported"
	// switch dot.Values["t"].(type) {
	// case int:
	// 	return "it's an int!"
	// case string:
	// 	return "it's a string!"
	// case float64:
	// 	return "it's a float64!"
	// case bool:
	// 	return "it's a bool!"
	// default:
	// 	return "it's something else"
	// }
}
