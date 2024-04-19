// +gotohelm:ignore=true
package redpanda

import (
	_ "embed"
	"encoding/json"

	"github.com/invopop/jsonschema"
)

//go:embed values.schema.json
var schemaBytes []byte

func JSONSchema() *jsonschema.Schema {
	var s jsonschema.Schema
	if err := json.Unmarshal(schemaBytes, &s); err != nil {
		panic(err)
	}
	return &s
}
