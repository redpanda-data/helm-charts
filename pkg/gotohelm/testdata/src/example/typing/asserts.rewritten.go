//go:build rewrites
package typing

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func typeTesting(dot *helmette.Dot) string {
	t := dot.Values["t"]
	tmp_tuple_1 := helmette.Compact2(helmette.TypeTest[string](t))
	ok_1 := tmp_tuple_1.T2
	tmp_tuple_2 := helmette.Compact2(helmette.AsIntegral[int](t))
	ok_2 := tmp_tuple_2.
		// } else if _, ok := t.(int); ok {
		T2
	tmp_tuple_3 := helmette.Compact2(

		// } else if _, ok := t.(float64); ok {
		helmette.AsNumeric(t))
	ok_3 := tmp_tuple_3.T2
	if ok_1 {
		return "it's a string!"
	} else if ok_2 {

		return "it's an int!"

	} else if ok_3 {
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
