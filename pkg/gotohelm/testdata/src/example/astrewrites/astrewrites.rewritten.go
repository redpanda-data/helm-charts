//go:build rewrites
package astrewrites

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
)

func ASTRewrites() []any {
	return []any{}
}

func mvrs() {
	m := map[string]int{}
	var a any = m

	{
		tmp_tuple_1 := helmette.Compact2(helmette.DictTest[string, int](m, "1"))
		y := tmp_tuple_1.T2
		x := tmp_tuple_1.T1
		_ = x
		_ = y
	}

	{
		tmp_tuple_2 := helmette.Compact2(helmette.TypeTest[map[string]int](a))
		y := tmp_tuple_2.T2
		x := tmp_tuple_2.T1
		_, _ = x, y
	}

	{
		tmp_tuple_3 := helmette.Compact2(helmette.TypeTest[map[string]int](a))
		x := tmp_tuple_3.T1
		_ = x
	}

	{
		tmp_tuple_4 := helmette.Compact2(helmette.TypeTest[map[string]int](a))
		x := tmp_tuple_4.T2
		_ = x
	}

	{
		_ = helmette.Compact2(helmette.TypeTest[map[string]int](a))
	}

	{
		tmp_tuple_6 := helmette.Compact3(mvr3())
		c := tmp_tuple_6.T3
		b := tmp_tuple_6.T2
		a := tmp_tuple_6.T1
		_, _, _ = a, b, c
	}

	{
		// Using a 3rd party type, with type aliasing to boot.
		m := map[string]corev1.Affinity{}
		tmp_tuple_7 := helmette.Compact2(helmette.DictTest[string, corev1.Affinity](m, ""))
		y := tmp_tuple_7.T2
		x := tmp_tuple_7.T1
		_, _ = x, y
	}
}

type mymap map[string]int

func dictTest() {
	m := mymap{}
	tmp_tuple_8 := helmette.Compact2(helmette.DictTest[string, int](m, ""))
	ok := tmp_tuple_8.T2
	_ = ok
}

func typeTest() {
	var m any = map[string]int{}
	tmp_tuple_9 := helmette.Compact2(helmette.TypeTest[map[string]string](m))
	ok := tmp_tuple_9.T2
	_ = ok
	_ = helmette.Compact2(helmette.TypeTest[map[string]int](m))
}

func ifHoisting() {
	m := map[string]int{"1": 1}
	tmp_tuple_11 := helmette.Compact2(helmette.DictTest[string, int](m, "2"))
	ok_1 := tmp_tuple_11.T2
	tmp_tuple_12 := helmette.Compact2(helmette.DictTest[string, int](m, "3"))
	ok_2 := tmp_tuple_12.T2
	tmp_tuple_13 := helmette.Compact2(helmette.DictTest[string, int](m, "4"))
	ok_3 := tmp_tuple_13.T2
	tmp_tuple_14 := helmette.Compact2(helmette.DictTest[string, int](m, "5"))
	ok_4 := tmp_tuple_14.T2
	if ok_1 {
	} else if ok_2 {
	} else if ok_3 {
	} else if ok_4 {
	} else {
	}
}

func mvr3() (float32, bool, int) {
	return 0, true, 3
}
