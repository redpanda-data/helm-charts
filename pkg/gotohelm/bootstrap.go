package gotohelm

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var (
	// bootstrapGo is the internal/bootstrap package but embedded so it can be
	// transpiled on the fly.
	//go:embed internal/bootstrap/*.go
	bootstrapGo embed.FS

	// shims the source [File] of _shims.tpl. It's set by the init function in
	// bootstrap.go.
	shims *File
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

func init() {
	// Oh yes. We transpile the bootstrap package when this package is first
	// loaded to generate _shims.tpl. It's a weird process but removes any
	// possibility of things getting out of sync.
	dir, _ := os.Getwd()

	// First, we always bind Dir to the working directory. It could be any
	// directory as we really just need absolute paths.
	// bootstrapGo is turned into an Overlay such that files get loaded from it
	// instead of go trying to find the package on disk.
	// The bootstrap package MUST NOT load any 3rd party libraries as the
	// loader will start complaining about the lack of a go.mod.
	pkgs, err := LoadPackages(&packages.Config{
		Dir:     dir,
		Overlay: fsToOverlay(&bootstrapGo, dir),
	}, filepath.Join(dir, "internal/bootstrap"))
	if err != nil {
		panic(err)
	}

	// Then we transpile the loaded package as we would any other.
	bootstrapChart, err := Transpile(pkgs[0])
	if err != nil {
		panic(err)
	}

	// The result is then shoved into a global variable so all further calls to
	// Transpile can make use of our shim/bootstrap layer.
	shims = bootstrapChart.Files[0]

	// Attach a foot of helpers written in raw gotpl that can't be expressed in
	// gotohelm.
	shims.Footer = renderManifest
}

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
