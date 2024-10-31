package connect

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelmUnitTest(t *testing.T) {
	out, err := exec.Command("helm", "unittest", ".").CombinedOutput()
	require.NoError(t, err, "failed to run helm unittest: %s", out)
}
