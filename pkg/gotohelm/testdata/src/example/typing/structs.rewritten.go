//go:build rewrites
package typing

type NestedEmbeds struct {
	WithEmbed
}

type Object struct {
	Key     string
	WithTag int `json:"with_tag"`
}

type WithEmbed struct {
	Object
	Exclude string  `json:"-"`
	Omit    *string `json:"Omit,omitempty"`
	Nilable *int
}

type JSONKeys struct {
	Value    string      `json:"val,omitempty"`
	Children []*JSONKeys `json:"childs,omitempty"`
}

func zeros() []any {
	var number *int
	var str *string
	var stru *Object

	return []any{
		Object{},
		WithEmbed{},
		number,
		str,
		stru,
	}
}

func nestedFieldAccess() string {
	x := JSONKeys{
		Children: []*JSONKeys{
			{
				Children: []*JSONKeys{
					{Value: "Hello!"},
				},
			},
		},
	}

	return x.Children[0].Children[0].Value
}

func settingFields() []string {
	var out NestedEmbeds

	out.WithEmbed = WithEmbed{Object: Object{Key: "foo"}}
	out.Object = Object{Key: "bar"}
	out.Key = "quux"

	return []string{
		out.Key,
		out.Object.Key,
		out.WithEmbed.Key,
	}
}

func compileMe() Object {
	return Object{
		Key: "foo",
	}
}

func alsoMe() WithEmbed {
	return WithEmbed{
		Object: Object{
			Key: "Foo",
		},
		Exclude: "Exclude",
	}
}
