package valuesutil

import (
	"bytes"
	"encoding/json"

	"github.com/invopop/jsonschema"
	schemavalidator "github.com/santhosh-tekuri/jsonschema/v5"
)

// Validate returns an error if instance is not considered valid by schema.
// Otherwise it returns nil.
func Validate(schema *jsonschema.Schema, instance any) error {
	c := schemavalidator.NewCompiler()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(schema); err != nil {
		return err
	}

	if err := c.AddResource(string(schema.ID), &buf); err != nil {
		return err
	}

	validator, err := c.Compile(string(schema.ID))
	if err != nil {
		return err
	}

	return validator.Validate(instance)
}

func GenerateSchema(instance any) *jsonschema.Schema {
	r := jsonschema.Reflector{
		ExpandedStruct:             true,
		DoNotReference:             true,
		RequiredFromJSONSchemaTags: true,
	}

	return r.Reflect(instance)
}

// RoundTripThrough round trips input through T. It may be used to understand
// how various types affect JSON marshalling or apply go's defaulting to an
// untyped value.
func RoundTripThrough[T any, K any](input K) (K, error) {
	through, err := UnmarshalInto[T](input)
	if err != nil {
		var zero K
		return zero, err
	}

	return UnmarshalInto[K](through)
}

// UnmarshalInto "converts" input into T by marshalling input to JSON and then
// unmarshalling into T.
func UnmarshalInto[T any](input any) (T, error) {
	var output T
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(input); err != nil {
		return output, err
	}

	if err := json.NewDecoder(&buf).Decode(&output); err != nil {
		return output, err
	}

	return output, nil
}
