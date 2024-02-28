package typing

import "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"

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

func Typing(dot *helmette.Dot) map[string]any {
	return map[string]any{
		"zeros": zeros(),
		// "settingFields":     settingFields(),
		"compileMe":         compileMe(),
		"typeTesting":       typeTesting(dot),
		"typeAssertions":    typeSwitching(dot),
		"typeSwitching":     typeSwitching(dot),
		"nestedFieldAccess": nestedFieldAccess(),
	}
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

// func settingFields() string {
// 	var out WithEmbed
//
// 	out.Object = Object{Key: "foo"}
// 	out.Object.Key = "bar"
// 	return out.Object.Key
// }

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
