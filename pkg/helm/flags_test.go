package helm_test

import (
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/stretchr/testify/require"
)

type Flags struct {
	NoWait        bool `flag:"wait"`
	NoWaitForJobs bool `flag:"no-wait-for-jobs"`
	NotAFlag      string
	StringFlag    string   `flag:"string-flag"`
	StringArray   []string `flag:"string-array"`
}

func TestToFlags(t *testing.T) {
	testCases := []struct {
		in  Flags
		out []string
	}{
		{
			in: Flags{},
			out: []string{
				"--wait=true",
				"--no-wait-for-jobs=false",
			},
		},
		{
			in: Flags{},
			out: []string{
				"--wait=true",
				"--no-wait-for-jobs=false",
			},
		},
		{
			in: Flags{
				StringFlag: "something",
			},
			out: []string{
				"--wait=true",
				"--no-wait-for-jobs=false",
				"--string-flag=something",
			},
		},
		{
			in: Flags{
				StringFlag:  "something",
				StringArray: []string{"1", "2", "3"},
			},
			out: []string{
				"--wait=true",
				"--no-wait-for-jobs=false",
				"--string-flag=something",
				"--string-array=1",
				"--string-array=2",
				"--string-array=3",
			},
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.out, helm.ToFlags(tc.in))
	}
}
