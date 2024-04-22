// +gotohelm:ignore=true
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
	// Schema for memory amount is a subset of api machinery resource quantity.
	//
	// Reference https://github.com/kubernetes/apimachinery/blob/0ee3e6150890f56b226a3fe5a95ba33b1b2bf7c7/pkg/api/resource/quantity.go#L35-L57
	//
	// The serialization format is:
	//
	// ```
	// <quantity>        ::= <signedNumber><suffix>
	//
	//	(Note that <suffix> may be empty, from the "" case in <decimalSI>.)
	//
	// <digit>           ::= 0 | 1 | ... | 9
	// <digits>          ::= <digit> | <digit><digits>
	// <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits>
	// <sign>            ::= "+" | "-"
	// <signedNumber>    ::= <number> | <sign><number>
	// <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI>
	// <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei
	//
	//	(International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)
	//
	// <decimalSI>       ::= m | "" | k | M | G | T | P | E
	//
	//	(Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)
	//
	// <decimalExponent> ::= "e" <signedNumber> | "E" <signedNumber>
	// ```
	schema.Pattern = "^[0-9]+(\\.[0-9]){0,1}(k|M|G|T|P|Ki|Mi|Gi|Ti|Pi)?$"
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

type HTTPAuthenticationMethod string

func (HTTPAuthenticationMethod) JSONSchemaExtend(s *jsonschema.Schema) {
	s.Enum = append(s.Enum, "none", "http_basic")
}

type KafkaAuthenticationMethod string

func (KafkaAuthenticationMethod) JSONSchemaExtend(s *jsonschema.Schema) {
	s.Enum = append(s.Enum, "sasl", "none", "mtls_identity")
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

func makeNullable(schema *jsonschema.Schema, keys ...string) {
	for _, key := range keys {
		prop, ok := schema.Properties.Get(key)
		if !ok {
			panic(fmt.Sprintf("missing field %q on %T", key, schema.Title))
		}
		schema.Properties.Set(key, &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				prop,
				{Type: "null"},
			},
		})
	}
}
