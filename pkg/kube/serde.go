package kube

import (
	"bufio"
	"bytes"
	"io"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

var (
	NewYAMLFrameReader = json.YAMLFramer.NewFrameReader
	NewYAMLFrameWriter = json.YAMLFramer.NewFrameWriter
)

// EncodeYAML calls [EncodeYAMLInto] with a [bytes.Buffer] and returns the
// resultant bytes.
func EncodeYAML(scheme *runtime.Scheme, objs ...Object) ([]byte, error) {
	var b bytes.Buffer
	if err := EncodeYAMLInto(NewYAMLFrameWriter(&b), scheme, objs...); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// EncodeYAMLInto encodes a slice of [Object]s into multi-document YAML and
// writes them to w.
// NOTE: .TypeMeta of all provided objects WILL BE SET BY THIS FUNCTION.
func EncodeYAMLInto(w io.Writer, scheme *runtime.Scheme, objs ...Object) error {
	if scheme == nil {
		scheme = clientscheme.Scheme
	}

	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)

	w = NewYAMLFrameWriter(w)

	for _, obj := range objs {
		gvk, err := apiutil.GVKForObject(obj, scheme)
		if err != nil {
			return errors.WithStack(err)
		}

		obj.GetObjectKind().SetGroupVersionKind(gvk)

		if err := serializer.Encode(obj, w); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// DecodeYAML calls [DecodeYAMLFrom] with a [bytes.Reader].
func DecodeYAML(manifest []byte, scheme *runtime.Scheme) ([]Object, error) {
	return DecodeYAMLFrom(bytes.NewReader(manifest), scheme)
}

// DecodeYAMLFrom decodes a multi-document YAML into a slice of concretely
// types [kube.Object]s.
// To appropriately decode, a scheme that's knowledgable of all the provided
// types must be provided. If none is provided, the scheme from Kubernetes' go
// client will be used.
func DecodeYAMLFrom(in io.Reader, scheme *runtime.Scheme) ([]Object, error) {
	if scheme == nil {
		scheme = clientscheme.Scheme
	}

	reader := yamlutil.NewYAMLReader(bufio.NewReader(in))
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	var objects []client.Object

	for {
		doc, err := reader.Read()
		if err == io.EOF {
			return objects, nil
		}

		if err != nil {
			return nil, err
		}

		obj, _, err := decoder.Decode(doc, nil, nil)
		if err != nil {
			// Special case to work around an issue with helm outputs. There can be empty YAML docs in the form:
			// ---
			// # Source: ....
			// ---
			// Parsing these docs as YAML will result in a nil value. If we
			// can't decode a k8s object and the YAML is otherwise parsed as
			// nil, skip the document.
			var x any
			if yaml.Unmarshal(doc, &x) == nil && x == nil {
				continue
			}
			return nil, err
		}

		objects = append(objects, obj.(Object))
	}
}
