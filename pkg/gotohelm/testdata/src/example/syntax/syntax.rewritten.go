//go:build rewrites
package syntax

import (
	"fmt"
	"math"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
)

const (
	AStrConst           = "1234"
	AnIntConst          = 1234
	ARationalFloatConst = 0.1
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

	// ParenExpr
	_ = (true)

	// SliceExpr
	slice := sliceExpr()

	// Ident
	_ = AStrConst           // A reference to a string constant
	_ = AnIntConst          // A reference to an int constant
	_ = ARationalFloatConst // A reference to a rational float constant

	// SelectorExpr
	_ = corev1.IPv4Protocol // A reference to an imported constant
	_ = math.E              // reference to an irrational float constant

	// TypeAssertExpr
	var x any
	_ = helmette.
		// _, _ = x.(int) // Numeric types will generate an error.
		Compact2(helmette.TypeTest[[]any](x))
	_ = helmette.Compact2(helmette.TypeTest[[]string](x))
	_ = helmette.Compact2(helmette.TypeTest[map[string]any](x))

	return map[string]any{
		"sliceExpr":       slice,
		"negativeNumbers": []int{-2, -4},
		"forExpr":         forExpr(10, Complex{Iterations: 5}),
		"binaryExprs":     binaryExprs(),
		"instance-method": instanceMethod(),
	}
}

type TestStruct struct {
	TestBoolean bool
	Mult        int
	SomeString  string
}

func (ts *TestStruct) MutateString(input string) {
	ts.SomeString = input
}

func (ts TestStruct) DoNotMutateString(input string) {
	ts.SomeString = input
}

func (ts *TestStruct) Double(input int) int {
	return input * 2
}

func (ts *TestStruct) Multiplayer(input int) int {
	return input * ts.Mult
}

func (ts *TestStruct) InstanceMethod() bool {
	return ts.TestBoolean
}

func (ts TestStruct) String(arg1, arg2 string) string {
	return fmt.Sprintf(ts.SomeString, arg1, arg2)
}

func instanceMethod() any {
	t := TestStruct{
		TestBoolean: true,
		Mult:        4,
		SomeString:  "%s and %s",
	}
	f := TestStruct{
		TestBoolean: false,
		Mult:        5,
	}

	f.MutateString("Change string")
	f.DoNotMutateString("do not change")
	return []any{
		t.InstanceMethod(),
		f.InstanceMethod(),
		t.Double(2),
		t.Double(4),
		t.Multiplayer(6),
		f.Multiplayer(6),
		t.String("one", "two"),
		f.SomeString == "Change string",
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
func binaryExprs() []bool {
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

	result := []bool{
		1 > 2,
		1 < 2,
		1 >= 2,
		1 <= 2,
		2 >= 2,
		2 <= 2,
	}

	return result
}

type Complex struct {
	Iterations int
}

func forExpr(iteration int, in Complex) [][]string {
	result := [][]string{}

	// ["0","1","2","3","4"]
	test := []string{}
	for i := 0; i < in.Iterations; i++ {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["0","1","2","3","4","5","6","7","8","9"]
	test = []string{}
	for i := 0; i < iteration; i++ {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["2","3","4","5","6","7","8","9"]
	test = []string{}
	for i := 2; i < iteration; i++ {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["2","3","4"]
	test = []string{}
	for i := 2; i < in.Iterations; i++ {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["2","4","6","8"]
	test = []string{}
	for i := 2; i < iteration; i += 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["2","4"]
	test = []string{}
	for i := 2; i < in.Iterations; i += 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["17","15","13","11"]
	test = []string{}
	for i := 17; i > iteration; i -= 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// ["17","15","13","11","9","7"]
	test = []string{}
	for i := 17; i > in.Iterations; i -= 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	// This for loop is noop as the condition should be greater than. The test array will be empty.
	test = []string{}
	for i := 17; i < iteration; i -= 2 {
		test = append(test, fmt.Sprintf("%d", i))
	}
	result = append(result, test)

	return result
}
