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
	testCases := []struct {
		Name   string
		Inputs []map[string]any
		Output map[string]any
	}{
		{
			"matching_key",
			[]map[string]any{
				{"label": "something-default", "none-overlapping-key": "test"},
				{"label": "second-will-not-overwrite"},
			},
			map[string]any{"label": "something-default", "none-overlapping-key": "test"},
		},
		{
			"multiple_key",
			[]map[string]any{
				{"label": "something-default"},
				{"label": "second-will-not-overwrite"},
				{"label": "last-overwritten-also-will-not-overwrite"},
			},
			map[string]any{"label": "something-default"},
		},
		{
			"empty_map_as_destination_first_map",
			[]map[string]any{
				{},
				{"label": "first-in-line"},
				{"label": "last-value-for-the-label-key"},
			},
			map[string]any{"label": "first-in-line"},
		},
		{
			"none_empty_map",
			[]map[string]any{
				{"key": "value"},
				{"label": "overwritten"},
				{"label": "last-overwritten"},
			},
			map[string]any{"label": "overwritten", "key": "value"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			require.Equal(t, tc.Output, helmette.Merge(tc.Inputs...))
		})
	}
}

func TestMergeBorrowedFromSprig(t *testing.T) {
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

	expectedBytes, err := json.Marshal(expected)
	require.NoError(t, err)

	dstBytes, err := json.Marshal(dict["dst"])
	require.NoError(t, err)

	answerBytes, err := json.Marshal(answer)
	require.NoError(t, err)

	assert.Equal(t, expectedBytes, []uint8(output))
	assert.Equal(t, expectedBytes, dstBytes)
	assert.Equal(t, expectedBytes, answerBytes)
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
