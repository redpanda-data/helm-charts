package testutil

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"
)

var retain = flag.Bool("retain", false, "if true, no clean up will be performed.")

// Retain returns the value of the -retain CLI flag. A value of true indicates
// that cleanup actions should be SKIPPED.
func Retain() bool {
	return *retain
}

// TempDir is wrapper around [testing.T.TempDir] that respects [Retain].
func TempDir(t *testing.T) string {
	t.Helper()
	if !Retain() {
		return t.TempDir()
	}
	dir, err := os.MkdirTemp(os.TempDir(), t.Name())
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return dir
}

// MaybeCleanup is helper to invoke `fn` within a [testing.T.Cleanup] closure
// only if [Retain] returns false.
func MaybeCleanup(t *testing.T, fn func()) {
	t.Cleanup(func() {
		if Retain() {
			return
		}
		fn()
	})
}

// Context returns a [context.Context] that will cancel 1s before the t's
// deadline.
func Context(t *testing.T) context.Context {
	ctx := context.Background()
	if timeout, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, timeout.Add(-time.Second))
		t.Cleanup(cancel)
	}
	return ctx
}

// Writer wraps a [testing.T] to implement [io.Writer] by utilizing
// [testing.T.Log].
type Writer struct {
	T *testing.T
}

func (w Writer) Write(p []byte) (int, error) {
	w.T.Log(string(p))
	return len(p), nil
}
