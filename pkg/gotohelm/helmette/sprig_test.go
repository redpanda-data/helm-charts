package helmette_test

import (
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
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
