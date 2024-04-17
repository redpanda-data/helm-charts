package valuesutil

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"slices"

	"github.com/invopop/jsonschema"
	"github.com/lucasjones/reggen"
	"k8s.io/utils/ptr"
)

// Generate generates a go values that is valid for the provided JSON schema.
// It may be used to "fuzz" or property test charts to ensure that all paths
// are appropriately explored and that the schema is well formed for the chart.
func Generate(rng *rand.Rand, s *jsonschema.Schema) any {
	g := generator{rng: rng, maxDepth: 15, defs: buildDefMap(s), skipDeprecated: true}
	return g.generate(0, s)
}

func buildDefMap(s *jsonschema.Schema) jsonschema.Definitions {
	mapping := map[string]*jsonschema.Schema{"#" + s.Anchor: s}
	for def, schema := range s.Definitions {
		mapping["#/$defs/"+def] = schema
	}
	return mapping
}

type generator struct {
	defs           jsonschema.Definitions
	maxDepth       int
	rng            *rand.Rand
	skipDeprecated bool
	reggenCache    map[string]*reggen.Generator
}

func (g *generator) generateRegex(pattern string) string {
	if g.reggenCache == nil {
		g.reggenCache = map[string]*reggen.Generator{}
	}
	if _, ok := g.reggenCache[pattern]; !ok {
		gen := must(reggen.NewGenerator(pattern))
		gen.SetSeed(int64(g.rng.Int()))
		g.reggenCache[pattern] = gen
	}
	return g.reggenCache[pattern].Generate(5)
}

func (g *generator) generate(depth int, s *jsonschema.Schema) any {
	if depth > g.maxDepth {
		panic("exceeded max depth")
	}

	serialized := string(must(json.Marshal(s)))

	switch {
	case s == nil:
		panic("Why is s nil?")

	case serialized == "true":
		return nil

	case s.Ref != "":
		if referred, ok := g.defs[s.Ref]; ok {
			return g.generate(depth, referred)
		}
		panic(fmt.Sprintf("unknown ref: %q", s.Ref))

	case s.DynamicRef != "":
		if referred, ok := g.defs[s.DynamicRef]; ok {
			return g.generate(depth, referred)
		}
		panic(fmt.Sprintf("unknown dynamic ref: %q", s.DynamicRef))

	case s.Const != nil:
		return s.Const

	case len(s.Enum) > 0:
		return pickone(g.rng, s.Enum)

	case len(s.OneOf) > 0:
		return g.generate(depth, pickone(g.rng, s.OneOf))

	case len(s.AnyOf) > 0:
		return g.generate(depth, pickone(g.rng, s.AnyOf))
	}

	switch s.Type {
	case "null":
		return nil

	case "boolean":
		// Flip a coin.
		return g.rng.Intn(2) == 0

	case "number":
		fallthrough

	case "integer":
		max, _ := s.Maximum.Int64()
		min, _ := s.Minimum.Int64()
		if max == 0 {
			max = math.MaxInt16
			// TODO this causes too many issues right now as setting a type to
			// int doesn't automatically result in the upper and lower
			// boundaries being applied.
			// max = math.MaxInt64
		}
		return int(min) + g.rng.Intn(int(max-min))

	case "string":
		if s.Pattern != "" {
			if s.MinLength != nil || s.MaxLength != nil {
				panic("unsupported pattern + min/maxlength")
			}
			return g.generateRegex(s.Pattern)
		}

		min := int(ptr.Deref(s.MinLength, 0))
		max := int(ptr.Deref(s.MaxLength, 10))

		var result string
		const alphabet = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()-_=+"
		for i := 0; i < min+g.rng.Intn(max-min); i++ {
			result += string(alphabet[g.rng.Intn(len(alphabet))])
		}
		return result

	case "array":
		min := int(ptr.Deref(s.MinItems, 0))
		max := int(ptr.Deref(s.MaxItems, 5))

		itemSchema := s.Items
		if itemSchema == nil {
			itemSchema = jsonschema.TrueSchema
		}

		items := []any{}
		for i := 0; i < min+g.rng.Intn(max-min); i++ {
			items = append(items, g.generate(depth, itemSchema))
		}
		return items

	case "object":
		val := map[string]any{}
		min := int(ptr.Deref(s.MinProperties, uint64(s.Properties.Len())))
		max := int(ptr.Deref(s.MaxProperties, uint64(s.Properties.Len()+5)))

		for _, key := range s.Required {
			schema, ok := s.Properties.Get(key)
			if !ok {
				// This might be okay in cases of AdditionalProperties or pattern properties...
				panic(fmt.Sprintf("missing required property %q", key))
			}

			if g.skipDeprecated && schema.Deprecated {
				continue
			}

			val[key] = g.generate(depth, schema)
		}

		var patterns []string
		var schemas []*jsonschema.Schema

		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			required := slices.Contains(s.Required, pair.Key)
			if required || (g.skipDeprecated && pair.Value.Deprecated) {
				continue
			}

			patterns = append(patterns, regexp.QuoteMeta(pair.Key))
			schemas = append(schemas, pair.Value)
		}

		for pattern, schema := range s.PatternProperties {
			if g.skipDeprecated && schema.Deprecated {
				continue
			}
			patterns = append(patterns, pattern)
			schemas = append(schemas, schema)
		}

		addlProps := string(must(json.Marshal(s.AdditionalProperties)))
		if addlProps != "null" && addlProps != "false" {
			if !(g.skipDeprecated && s.AdditionalProperties.Deprecated) {
				patterns = append(patterns, `\w+`)
				schemas = append(schemas, s.AdditionalProperties)
			}
		}

		// If there are no optional properties to add, bail early or if we're
		// starting to get too "deep" progressively decrease the likelihood of
		// going even deeper.
		if len(patterns) == 0 || depth+g.rng.Intn(g.maxDepth) >= g.maxDepth {
			return val
		}

		for i := len(val); i < min+g.rng.Intn(max-min); i++ {
			j := g.rng.Intn(len(patterns))

			schema := schemas[j]
			pattern := patterns[j]

			key := g.generateRegex(pattern)
			val[key] = g.generate(depth+1, schema)
		}

		return val

	default:
		panic(fmt.Sprintf("unhandled schema: %s", serialized))
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func pickone[T any](rng *rand.Rand, l []T) T {
	return l[rng.Intn(len(l))]
}
