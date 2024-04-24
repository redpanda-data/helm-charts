package redpanda_test

import (
	"math"
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/stretchr/testify/require"
)

func TestSIToBytes(tt *testing.T) {
	// NB: This intentionally uses math.Pow as `SIToBytes` relies on
	// calculating values by chaining multiplication.
	k := int64(1000)
	m := int64(math.Pow(1000, 2))
	g := int64(math.Pow(1000, 3))
	t := int64(math.Pow(1000, 4))
	p := int64(math.Pow(1000, 5))

	ki := int64(1024)
	mi := int64(math.Pow(1024, 2))
	gi := int64(math.Pow(1024, 3))
	ti := int64(math.Pow(1024, 4))
	pi := int64(math.Pow(1024, 5))

	cases := []struct {
		In   string
		Want int64
	}{
		{In: "0", Want: 0},
		{In: "1", Want: 1},
		{In: "1024", Want: 1024},
		{In: "1k", Want: 1 * k},
		{In: "1Ki", Want: 1 * ki},
		{In: "1M", Want: 1 * m},
		{In: "1Mi", Want: 1 * mi},
		{In: "1G", Want: 1 * g},
		{In: "1Gi", Want: 1 * gi},
		{In: "1024Gi", Want: 1024 * gi},
		{In: "1024G", Want: 1024 * g},
		{In: "1T", Want: 1 * t},
		{In: "1Ti", Want: 1 * ti},
		{In: "1024Ti", Want: 1024 * ti},
		{In: "1024T", Want: 1024 * t},
		{In: "1Pi", Want: 1 * pi},
		{In: "1P", Want: 1 * p},
		{In: "2.5Gi", Want: 2684354560},
	}

	for _, c := range cases {
		tt.Run(c.In, func(t *testing.T) {
			require.Equalf(tt, int(c.Want), redpanda.SIToBytes(c.In), "%q is %d bytes", c.In, c.Want)
		})
	}
}
