package redpanda_test

import (
	"math"
	"strings"
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/stretchr/testify/require"
)

func TestSIToBytes(t *testing.T) {
	// NB: This intentionally uses math.Pow as `SIToBytes` relies on
	// calculating values by chaining multiplication.
	k := 1000
	m := int(math.Pow(1000, 2))
	g := int(math.Pow(1000, 3))

	ki := 1024
	mi := int(math.Pow(1024, 2))
	gi := int(math.Pow(1024, 3))

	cases := []struct {
		In   string
		Want int
	}{
		{In: "0", Want: 0},
		{In: "1", Want: 1},
		{In: "1024", Want: 1024},
		{In: "1K", Want: 1 * k},
		{In: "1Ki", Want: 1 * ki},
		{In: "1M", Want: 1 * m},
		{In: "1Mi", Want: 1 * mi},
		{In: "1G", Want: 1 * g},
		{In: "1Gi", Want: 1 * gi},
		{In: "1024Gi", Want: 1024 * gi},
		{In: "1024G", Want: 1024 * g},
	}

	for _, c := range cases {
		for _, variation := range []string{
			c.In,
			strings.ToLower(c.In),
			strings.ToUpper(c.In),
		} {
			require.Equalf(t, c.Want, redpanda.SIToBytes(variation), "%q is %d bytes", variation, c.Want)
		}
	}
}
