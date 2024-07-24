package helm

import (
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/cli/values"
)

// MergeYAMLValues uses helm's values package to merge a collection of YAML
// values in accordance with helm's merging logic.
// Sadly, their merging logic is not exported nor can it accept raw JSON/YAML
// so we dump files on disk.
func MergeYAMLValues(tempDir string, vs ...[]byte) (map[string]any, error) {
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	var opts values.Options
	for i, v := range vs {
		file, err := os.CreateTemp(tempDir, fmt.Sprintf("values-%d.yaml", i))
		if err != nil {
			return nil, err
		}

		if _, err := file.Write(v); err != nil {
			return nil, err
		}

		if err := file.Close(); err != nil {
			return nil, err
		}
		opts.ValueFiles = append(opts.ValueFiles, file.Name())
	}

	return opts.MergeValues(nil)
}
