package syntax

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
	_ = []int{1, 2, 3}[:]
	_ = []int{1, 2, 3}[1:]
	_ = []int{1, 2, 3}[:2]
	_ = []int{1, 2, 3}[1:2]
	_ = []int{1, 2, 3}[1:2:3]
	_ = "1234"[:]
	_ = "1234"[1:]
	_ = "1234"[:2]
	_ = "1234"[1:2]

	return map[string]any{}
}

// binaryExprs are a bit tricky because we need to care about the types beyond
// the syntax. It get's its own function because it's so expansive.
func binaryExprs() {
	// untyped ints
	_ = 1 * 1
	_ = 1 + 1
	_ = 1 - 1
	_ = 1 / 1
	_ = 1 == 1
	_ = 1 != 1
	// typed ints
	_ = int(1) * int(1)
	_ = int(1) + int(1)
	_ = int(1) - int(1)
	_ = int(1) / int(1)
	_ = int(1) == int(1)
	_ = int(1) != int(1)
	// int32s
	_ = int32(1) * int32(1)
	_ = int32(1) + int32(1)
	_ = int32(1) - int32(1)
	_ = int32(1) / int32(1)
	_ = int32(1) == int32(1)
	_ = int32(1) != int32(1)
	// int64s
	_ = int64(1) * int64(1)
	_ = int64(1) + int64(1)
	_ = int64(1) - int64(1)
	_ = int64(1) / int64(1)
	_ = int64(1) == int64(1)
	_ = int64(1) != int64(1)
	// Maps
	_ = map[string]any{} == nil
	_ = map[string]any{} != nil
	// TODO strings
	// TODO floats
}
