//+gotohelm:ignore=true
package redpanda

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type ImageTag string

func (ImageTag) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Pattern = `^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$|^$`
}

type ImageRepository string

func (ImageRepository) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Pattern = "^[a-z0-9-_/.]+$"
}

type MemoryAmount string

func (MemoryAmount) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Pattern = "^[0-9]+(\\.[0-9]){0,1}(k|M|G|Ki|Mi|Gi)$"
}

type IssuerRefKind string

func (IssuerRefKind) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Enum = append(schema.Enum, "ClusterIssuer", "Issuer")
}

type ExternalListeners[T any] map[string]T

func (ExternalListeners[T]) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.PatternProperties = map[string]*jsonschema.Schema{
		`^[A-Za-z_][A-Za-z0-9_]*$`: schema.AdditionalProperties,
	}
	minProps := uint64(1)
	schema.MinProperties = &minProps
	schema.AdditionalProperties = nil
}

func deprecate(schema *jsonschema.Schema, keys ...string) {
	for _, key := range keys {
		prop, ok := schema.Properties.Get(key)
		if !ok {
			panic(fmt.Sprintf("missing field %q on %T", key, schema.Title))
		}
		prop.Deprecated = true
	}
}
