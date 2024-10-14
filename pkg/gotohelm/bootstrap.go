package gotohelm

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/tools/go/packages"
)

// renderManifest is a helper function to call and render the results of a
// gotohelm function as a Kubernetes manifest. It handles nil checking and
// rendering either slices or individual manifests. It additionally contains a
// bit of extra logic to cut the `status` and `creationTimestamp` fields out of
// manifests before rendering them. Such fields have been reported to cause
// issues with tools such as ArgoCD (See #1458). Removal is done at this level
// to avoid breaking the invariant that gotohelm produces templates that are
// equivalent to their go source (which include .Status and
// .CreationTimestamp).
//
// Usage:
//
//	{{- include "_shims.render-manifest" (list "template.ToRender" .) -}}
const renderManifest = `{{- define "_shims.render-manifest" -}}
{{- $tpl := (index . 0) -}}
{{- $dot := (index . 1) -}}
{{- $manifests := (get ((include $tpl (dict "a" (list $dot))) | fromJson) "r") -}}
{{- if not (typeIs "[]interface {}" $manifests) -}}
{{- $manifests = (list $manifests) -}}
{{- end -}}
{{- range $_, $manifest := $manifests -}}
{{- if ne (toJson $manifest) "null" }}
---
{{toYaml (unset (unset $manifest "status") "creationTimestamp")}}
{{- end -}}
{{- end -}}
{{- end -}}
`

// bootstrapGo is the internal/bootstrap package but embedded so it can be
// transpiled on the fly.
//
//go:embed internal/bootstrap/*.go
var bootstrapGo embed.FS

// transpileBootstrap transpiles the internal/bootstrap package of gotohelm on
// the fly with the notable exception that it does not kickoff the
// transpilation of the bootstrap package.
//
// It is wrapped in a [sync.OnceValues] to amortize any follow up calls.
var transpileBootstrap = sync.OnceValues(func() (*File, error) {
	// NB: Transpilation is ON DEMAND. Not in init().
	// This allows other packages/binaries to import gotohelm without needing
	// to have the `go` binary available (packages.Load shells out to `go
	// list`).
	// In the future, it may make more sense to artificially inject the
	// bootstrap package's files via an overlay rather than making it a
	// separate build step or something of that nature. For now,
	// sync.OnceValues is an okay middle ground.

	// Otherwise, yep, we transpile the bootstrap package on the fly.
	//
	// It's a weird process but removes any
	// possibility of things getting out of sync.
	dir, _ := os.Getwd()

	// We can reasonably convince packages.Load to read from the embedded FS.
	// Though it seems likely this isn't 100% reliable as we always execute
	// gotohelm from the root of this repository.
	pkgs, err := LoadPackages(&packages.Config{
		Dir:     dir,
		Overlay: fsToOverlay(&bootstrapGo, dir),
	}, filepath.Join(dir, "internal/bootstrap"))
	if err != nil {
		return nil, err
	}

	// We call the private transpile method which doesn't bundle the _shims.tpl
	// into the final chart.
	bootstrapChart, err := transpile(pkgs[0])
	if err != nil {
		return nil, err
	}

	shims := bootstrapChart.Files[0]

	// Attach a foot of helpers written in raw gotpl that can't be expressed in
	// gotohelm.
	shims.Footer = renderManifest

	return shims, nil
})

// fsToOverlay translates an [embed.FS] into a map[string][]byte suitable for
// use in [packages.Config.Overlay].
func fsToOverlay(fsys *embed.FS, prefix string) map[string][]byte {
	overlay := map[string][]byte{}
	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		contents, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}
		overlay[filepath.Join(prefix, path)] = contents
		return nil
	}); err != nil {
		panic(err)
	}
	return overlay
}
