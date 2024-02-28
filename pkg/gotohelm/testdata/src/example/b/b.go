package b

func Constant() string {
	return "foo"
}

func ConfigMap() map[string]any {
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"name":       Constant(),
	}
}
