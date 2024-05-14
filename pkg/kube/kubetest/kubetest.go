package kubetest

import (
	"os/exec"
	"testing"
	"time"

	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const controlPlaneVersion = "1.30.x"

// NewEnv starts a local kubernetes control plane via [envtest.Environment] and
// returns a [kube.Ctl] to access it. The provided [testing.T] will be used to
// shutdown the control plane at the end of the test.
func NewEnv(t *testing.T) *kube.Ctl {
	// TODO: Would be nice to instead just import setup-envtest but the package
	// isn't exactly friendly to be used as a library. Alternatively, we could
	// use nix to provide the etcd and kubeapi-server binaries as that's all
	// setup-envtest does.
	if _, err := exec.LookPath("setup-envtest"); err != nil {
		t.Fatalf("setup-envtest not found in $PATH. Did you forget nix develop?")
	}

	stdout, err := exec.Command("setup-envtest", "use", controlPlaneVersion, "-p", "path").CombinedOutput()
	require.NoError(t, err)

	env := envtest.Environment{
		BinaryAssetsDirectory:    string(stdout),
		ControlPlaneStartTimeout: 30 * time.Second,
		ControlPlaneStopTimeout:  30 * time.Second,
	}

	cfg, err := env.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, env.Stop())
	})

	ctl, err := kube.FromRESTConfig(cfg)
	require.NoError(t, err)

	return ctl
}
