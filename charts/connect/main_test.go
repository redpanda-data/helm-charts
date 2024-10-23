package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestHelmUnitTest(t *testing.T) {
	cmd := exec.Command("helm", "unittest", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run helm unittest: %s", err)
	}
}
