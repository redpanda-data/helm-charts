package helmette_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDig(t *testing.T) {
	values := map[string]any{
		"k1": "v1",
		"k2": map[string]any{
			"k3": "v3",
			"k4": map[string]any{
				"k5": "v5",
			},
		},
	}

	require.Equal(t, "v1", helmette.Dig(values, "fallback", "k1"))
	require.Equal(t, "v3", helmette.Dig(values, "fallback", "k2", "k3"))
	require.Equal(t, "v5", helmette.Dig(values, "fallback", "k2", "k4", "k5"))
}

func TestDefault(t *testing.T) {
	require.Equal(t, 10, helmette.Default(10, 0))
	require.Equal(t, 20, helmette.Default(10, 20))

	require.Equal(t, "bar", helmette.Default("bar", ""))
	require.Equal(t, "foo", helmette.Default("bar", "foo"))

	require.Equal(t, map[string]any{"default": true}, helmette.Default(map[string]any{"default": true}, nil))
	require.Equal(t, map[string]any{"default": true}, helmette.Default(map[string]any{"default": true}, map[string]any{}))
	require.Equal(t, map[string]any{"default": false}, helmette.Default(map[string]any{"default": true}, map[string]any{"default": false}))
}

func TestMerge(t *testing.T) {
	t.Run("matching_key", func(t *testing.T) {
		require.Equal(t, map[string]any{"label": "something-default", "none-overlapping-key": "test"},
			helmette.Merge(
				map[string]any{"label": "something-default", "none-overlapping-key": "test"},
				map[string]any{"label": "overwritten"}))
	})

	t.Run("multiple_key", func(t *testing.T) {
		require.Equal(t, map[string]any{"label": "something-default"},
			helmette.Merge(
				map[string]any{"label": "something-default"},
				map[string]any{"label": "overwritten"},
				map[string]any{"label": "last-overwritten"}))
	})

	t.Run("empty_map_as_destination", func(t *testing.T) {
		dest := make(map[string]string)
		helmette.Merge(
			dest,
			map[string]string{"label": "first-in-line"},
			map[string]string{"label": "last-value-for-the-label-key"})
		require.Equal(t, map[string]string{"label": "first-in-line"}, dest)
	})

	t.Run("multiple_keys", func(t *testing.T) {
		dest := make(map[string]string)
		helmette.Merge(
			dest,
			map[string]string{"label": "first-in-line", "a": "test"},
			map[string]string{"label": "last-value-for-the-label-key", "b": "test"})
		require.Equal(t, map[string]string{"label": "first-in-line", "a": "test", "b": "test"}, dest)
	})

	t.Run("none_empty_map", func(t *testing.T) {
		noneEmpty := map[string]any{"key": "value"}
		helmette.Merge(
			noneEmpty,
			map[string]any{"label": "overwritten"},
			map[string]any{"label": "last-overwritten"})
		require.Equal(t, map[string]any{"label": "overwritten", "key": "value"}, noneEmpty)
	})

	t.Run("copy_of_sprig_merge_test_case", func(t *testing.T) {
		dict := map[string]any{
			"src2": map[string]any{
				"h": 10,
				"i": "i",
				"j": "j",
			},
			"src1": map[string]any{
				"a": 1,
				"b": 2,
				"d": map[string]any{
					"e": "four",
				},
				"g": []int{6, 7},
				"i": "aye",
				"j": "jay",
				"k": map[string]any{
					"l": false,
				},
			},
			"dst": map[string]any{
				"a": "one",
				"c": 3,
				"d": map[string]any{
					"f": 5,
				},
				"g": []int{8, 9},
				"i": "eye",
				"k": map[string]any{
					"l": true,
				},
			},
		}
		tpl := `{{merge .dst .src1 .src2 | toRawJson }}`
		output, err := runRawTemplate(tpl, dict)
		require.NoError(t, err)

		ts := make(map[string]any)

		err = json.Unmarshal([]byte(output), &ts)
		require.NoError(t, err)

		expected := map[string]any{
			"a": "one", // key overridden
			"b": 2,     // merged from src1
			"c": 3,     // merged from dst
			"d": map[string]any{ // deep merge
				"e": "four",
				"f": 5,
			},
			"g": []int{8, 9}, // overridden - arrays are not merged
			"h": 10,          // merged from src2
			"i": "eye",       // overridden twice
			"j": "jay",       // overridden and merged
			"k": map[string]any{
				"l": true, // overridden
			},
		}

		answer := helmette.Merge(map[string]any{
			"a": "one",
			"c": 3,
			"d": map[string]any{
				"f": 5,
			},
			"g": []int{8, 9},
			"i": "eye",
			"k": map[string]any{
				"l": true,
			},
		}, dict["src1"].(map[string]any), dict["src2"].(map[string]any))

		assert.Equal(t, changeAnyIntToFloat64(expected), ts)
		assert.Equal(t, expected, dict["dst"])
		assert.Equal(t, expected, answer)
	})
}

func changeAnyIntToFloat64(input map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range input {
		result[k] = v
		if intValue, ok := v.(int); ok {
			result[k] = float64(intValue)
		}
		if mapValue, ok := v.(map[string]any); ok {
			result[k] = changeAnyIntToFloat64(mapValue)
		}
		if arrayValue, ok := v.([]int); ok {
			var newArray []any
			for _, i := range arrayValue {
				newArray = append(newArray, float64(i))
			}
			result[k] = newArray
		}
	}
	return result
}

func runRawTemplate(tpl string, vars any) (string, error) {
	fmap := sprig.TxtFuncMap()
	t := template.Must(template.New("test").Funcs(fmap).Parse(tpl))
	var b bytes.Buffer
	err := t.Execute(&b, vars)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
