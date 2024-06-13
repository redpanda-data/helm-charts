package valuesutil

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/require"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"k8s.io/utils/ptr"
)

// TestGenerate is both an example of how to utilize [Generate] with
// [quick.Check] and a property test for [Generate]. It's a bit mind bending.
func TestGenerate(t *testing.T) {
	configs := []*quick.Config{
		// 1st property: Given a schema describing JSON schemas, we can generate
		// valid JSON schemas.
		{Values: func(args []reflect.Value, rng *rand.Rand) {
			args[0] = reflect.ValueOf(metaSchema)
			args[1] = reflect.ValueOf(Generate(rng, metaSchema))
		}},
		// 2st property: For any valid JSON schema, we can generate a value that is
		// valid. NB: We're relying on our 1st property to ensure that this
		// property always receives valid JSONSchemas.
		{Values: func(args []reflect.Value, rng *rand.Rand) {
			schema, _ := UnmarshalInto[*jsonschema.Schema](Generate(rng, metaSchema))
			args[0] = reflect.ValueOf(schema)
			args[1] = reflect.ValueOf(Generate(rng, metaSchema))
		}},
	}

	for _, cfg := range configs {
		if err := quick.Check(func(schema *jsonschema.Schema, instance any) bool {
			return Validate(schema, instance) == nil
		}, cfg); err != nil {
			err := err.(*quick.CheckError)
			schema := err.In[0].(*jsonschema.Schema)
			instance := err.In[1]

			t.Logf("Schema: %#v", schema)
			t.Logf("Invalid Generated Value: %#v", instance)
			require.NoError(t, err)
		}
	}
}

// TestGenerateDeterministic asserts that [Generate] is deterministic and
// produces the same value when given a particular seed.
func TestGenerateDeterministic(t *testing.T) {
	f := func(seed int64) any {
		return Generate(rand.New(rand.NewSource(seed)), metaSchema)
	}
	require.NoError(t, quick.CheckEqual(f, f, &quick.Config{}))
}

// FuzzGenerate is effectively equivalent to [TestGenerate]'s 1st property with
// the added benefit that it can save failures as regression cases to
// ./testdata.
func FuzzGenerate(f *testing.F) {
	f.Add(time.Now().UnixNano())
	f.Fuzz(func(t *testing.T, seed int64) {
		rng := rand.New(rand.NewSource(seed))
		instance := Generate(rng, metaSchema)
		require.NoError(t, Validate(metaSchema, instance))
	})
}

// metaSchema is a jsonschema that describes JSON schemas. It's used instead of
// loading the actual JSONSchema metaschema because [Generate] only implements
// a subset of features and the metaschema isn't fully complete as it allows
// specifying fields like maxProperties on arrays.
var metaSchema = &jsonschema.Schema{
	Definitions: jsonschema.Definitions{
		"any": &jsonschema.Schema{Const: true},
		"null": &jsonschema.Schema{
			Type:     "object",
			Required: []string{"type"},
			Properties: Props(map[string]*jsonschema.Schema{
				"type": {Const: "null"},
			}),
		},
		"integer": &jsonschema.Schema{
			Type:     "object",
			Required: []string{"type"},
			Properties: Props(map[string]*jsonschema.Schema{
				"type":    {Const: "integer"},
				"minimum": {Type: "integer", Minimum: json.Number("0"), Maximum: json.Number("1000")},
				"maximum": {Type: "integer", Minimum: json.Number("1001"), Maximum: json.Number("10000")},
			}),
		},
		"boolean": &jsonschema.Schema{
			Type:     "object",
			Required: []string{"type"},
			Properties: Props(map[string]*jsonschema.Schema{
				"type": {Const: "boolean"},
			}),
		},
		"string": &jsonschema.Schema{
			Type:     "object",
			Required: []string{"type"},
			Properties: Props(map[string]*jsonschema.Schema{
				"type":      {Const: "string"},
				"minLength": {Type: "integer", Minimum: json.Number("0"), Maximum: json.Number("3")},
				"maxLength": {Type: "integer", Minimum: json.Number("4"), Maximum: json.Number("7")},
			}),
		},
		"scalar": &jsonschema.Schema{
			AnyOf: []*jsonschema.Schema{
				{Ref: "#/$defs/any"},
				{Ref: "#/$defs/boolean"},
				{Ref: "#/$defs/integer"},
				{Ref: "#/$defs/null"},
				{Ref: "#/$defs/string"},
			},
		},
		"array": &jsonschema.Schema{
			Type:     "object",
			Required: []string{"type", "items"},
			Properties: Props(map[string]*jsonschema.Schema{
				"type": {Const: "array"},
				"items": {AnyOf: []*jsonschema.Schema{
					{Ref: "#/$defs/array"},
					{Ref: "#/$defs/object"},
					{Ref: "#/$defs/scalar"},
				}},
				"minItems": {Type: "integer", Minimum: json.Number("0"), Maximum: json.Number("3")},
				"maxItems": {Type: "integer", Minimum: json.Number("4"), Maximum: json.Number("7")},
			}),
		},
		"object": &jsonschema.Schema{
			Type:     "object",
			Required: []string{"type", "properties"},
			Properties: Props(map[string]*jsonschema.Schema{
				"type": {Const: "object"},
				"properties": {
					Type:          "object",
					MaxProperties: ptr.To(uint64(10)),
					MinProperties: ptr.To(uint64(0)),
					AdditionalProperties: &jsonschema.Schema{
						AnyOf: []*jsonschema.Schema{
							{Ref: "#/$defs/array"},
							{Ref: "#/$defs/object"},
							{Ref: "#/$defs/scalar"},
						},
					},
				},
				// TODO get additionalProperties and patternProperties working.
				// "patternProperties": {
				// 	Type:          "object",
				// 	MaxProperties: ptr.To(uint64(10)),
				// 	MinProperties: ptr.To(uint64(0)),
				// 	PatternProperties: map[string]*jsonschema.Schema{
				// 		`(\.|\\w|\\d){3,5}`: {
				// 			AnyOf: []*jsonschema.Schema{
				// 				{Ref: "#/$defs/array"},
				// 				{Ref: "#/$defs/object"},
				// 				{Ref: "#/$defs/scalar"},
				// 			},
				// 		},
				// 	},
				// },
				// "additionalProperties": {
				// 	Type: "object",
				// 	OneOf: []*jsonschema.Schema{
				// 		{Ref: "#/$defs/array"},
				// 		{Ref: "#/$defs/object"},
				// 		{Ref: "#/$defs/scalar"},
				// 	},
				// },
			}),
		},
	},
	AnyOf: []*jsonschema.Schema{
		{Ref: "#/$defs/object"},
	},
}

// Props is a helper for constructing inline OrderedMaps.
func Props(init map[string]*jsonschema.Schema) *orderedmap.OrderedMap[string, *jsonschema.Schema] {
	props := jsonschema.NewProperties()
	for key, value := range init {
		props.Set(key, value)
	}
	return props
}
