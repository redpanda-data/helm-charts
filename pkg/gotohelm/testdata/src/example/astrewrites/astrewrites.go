package astrewrites

import (
	corev1 "k8s.io/api/core/v1"
)

func ASTRewrites() []any {
	return []any{}
}

func mvrs() {
	m := map[string]int{}
	var a any = m

	{
		x, y := m["1"]
		_ = x
		_ = y
	}

	{
		x, y := a.(map[string]int)
		_, _ = x, y
	}

	{
		x, _ := a.(map[string]int)
		_ = x
	}

	{
		_, x := a.(map[string]int)
		_ = x
	}

	{
		_, _ = a.(map[string]int)
	}

	{
		a, b, c := mvr3()
		_, _, _ = a, b, c
	}

	{
		// Using a 3rd party type, with type aliasing to boot.
		m := map[string]corev1.Affinity{}
		x, y := m[""]
		_, _ = x, y
	}
}

type mymap map[string]int

func dictTest() {
	m := mymap{}
	_, ok := m[""]
	_ = ok
}

func typeTest() {
	var m any = map[string]int{}

	_, ok := m.(map[string]string)
	_ = ok

	_, _ = m.(map[string]int)
}

func ifHoisting() {
	m := map[string]int{"1": 1}

	if _, ok := m["2"]; ok {
	} else if _, ok := m["3"]; ok {
	} else if _, ok := m["4"]; ok {
	} else if _, ok := m["5"]; ok {
	} else {
	}
}

func mvr3() (float32, bool, int) {
	return 0, true, 3
}
