package mutability

import "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"

type Values struct {
	Name       string
	Labels     map[string]string
	SubService *SubService
}

type SubService struct {
	Name   string
	Labels map[string]string
}

func Mutability() map[string]any {
	var v Values
	v.Labels = map[string]string{}
	v.SubService = &SubService{}
	v.SubService.Labels = map[string]string{}

	v.SubService.Name = "Hello!"
	v.SubService.Labels["hello"] = "world"

	return map[string]any{
		"values": helmette.MustFromJSON(helmette.MustToJSON(v)),
	}
}
