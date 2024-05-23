package syntax

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

const (
	AStrConst  = "1234"
	AnIntConst = 1234
)

// Syntax only returns an empty map but it contains a variety of go syntax to
// assert that the transpiler doesn't crash upon seeing it.
// Notably: Syntax DOES NOT check for correctness.
func Syntax() map[string]any {
	// BasicLits
	_ = 0
	_ = ""
	_ = map[string]any{}
	_ = []int{}
	_ = ``
	_ = true
	_ = false

	// builtins
	_ = len("")
	_ = len([]int{})
	_ = len(map[string]any{})

	// BinaryExprs
	binaryExprs()

	// ParenExpr
	_ = (true)

	// SliceExpr
	slice := sliceExpr()

	// Ident
	_ = AStrConst  // A reference to a string constant
	_ = AnIntConst // A reference to an int constant

	// SelectorExpr
	_ = corev1.IPv4Protocol // A reference to an imported constant

	// TypeAssertExpr
	var x any
	_, _ = x.(int)
	_, _ = x.([]any)
	_, _ = x.([]string)
	_, _ = x.(map[string]any)

	return map[string]any{
		"sliceExpr": slice,
		"negativeNumbers": []int{-2, -4},
		"forExpr":   forExpr(10),
	}
}

func sliceExpr() map[string]any {
	_ = []int{1, 2, 3}[:]
	_ = []int{1, 2, 3}[1:]
	_ = []int{1, 2, 3}[:2]
	_ = []int{1, 2, 3}[1:2]
	_ = []int{1, 2, 3}[1:2:3]
	_ = "1234"[:]
	_ = "1234"[1:]
	_ = "1234"[:2]
	_ = "1234"[1:2]
	s := "abcd"
	_ = s[:len(s)-1]

	return workingWithString()
}

func workingWithString() map[string]any {
	amount := "2.5Gi"
	unit := string(amount[len(amount)-1])

	savedUnit := unit
	amount = amount[:len(amount)-1]

	if unit == "i" {
		// TODO string + string not implemented.
		unit = fmt.Sprintf("%s%s", amount[len(amount)-1:], unit)
		amount = amount[:len(amount)-1]
	}

	return map[string]any{
		"unit":          unit,
		"amount":        amount,
		"unitIsEqual":   unit == "Gi",
		"lastCharacter": savedUnit == "i",
	}
}

// binaryExprs are a bit tricky because we need to care about the types beyond
// the syntax. It get's its own function because it's so expansive.
func binaryExprs() {
	// untyped ints
	_ = 1 * 1
	_ = 1 + 1
	_ = 1 - 1
	_ = 1 / 1
	_ = 1 % 1
	_ = 1 == 1
	_ = 1 != 1
	// typed ints
	_ = int(1) * int(1)
	_ = int(1) + int(1)
	_ = int(1) - int(1)
	_ = int(1) / int(1)
	_ = int(1) % int(1)
	_ = int(1) == int(1)
	_ = int(1) != int(1)
	// int32s
	_ = int32(1) * int32(1)
	_ = int32(1) + int32(1)
	_ = int32(1) - int32(1)
	_ = int32(1) / int32(1)
	_ = int32(1) % int32(1)
	_ = int32(1) == int32(1)
	_ = int32(1) != int32(1)
	// int64s
	_ = int64(1) * int64(1)
	_ = int64(1) + int64(1)
	_ = int64(1) - int64(1)
	_ = int64(1) % int64(1)
	_ = int64(1) / int64(1)
	_ = int64(1) == int64(1)
	_ = int64(1) != int64(1)
	// Maps
	_ = map[string]any{} == nil
	_ = map[string]any{} != nil
	// TODO strings
	// TODO floats
}

func forExpr(interation int) [][]string {
	result := [][]string{}

	// ["0","1","2","3","4","5","6","7","8","9"]
	test := []string{}
	for i := 0; i < interation; i++ {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["2","3","4","5","6","7","8","9"]
	test = []string{}
	for i := 2; i < interation; i++ {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["2","4","6","8"]
	test = []string{}
	for i := 2; i < interation; i += 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["17","15","13","11"]
	test = []string{}
	for i := 17; i > interation; i -= 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// This for loop is noop as the condition should be greater than. The test array will be empty.
	test = []string{}
	for i := 17; i < interation; i -= 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	return result
}
