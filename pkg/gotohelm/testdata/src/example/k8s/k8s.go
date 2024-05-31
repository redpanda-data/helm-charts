package k8s

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

type Values struct {
	Quantity *resource.Quantity
}

func K8s(dot *helmette.Dot) map[string]any {
	return map[string]any{
		"Objects": []metav1.Object{
			pod(),
			pdb(),
			service(),
		},
		// intstr's are special cased because they have an... interesting
		// JSON/YAML mapping.
		"intstr": []intstr.IntOrString{
			intstr.FromInt(10),
			intstr.FromInt32(11),
			intstr.FromString("12"),
		},
		"ptr.Deref": []any{
			ptr.Deref(ptr.To(3), 4),
			ptr.Deref(nil, 3),
			ptr.Deref(ptr.To(""), "oh?"),
		},
		"ptr.To": []any{
			ptr.To("hello"),
			ptr.To(0),
			ptr.To(map[string]string{}),
		},
		"ptr.Equal": []bool{
			ptr.Equal[int](nil, nil),
			ptr.Equal(nil, ptr.To(3)),
			ptr.Equal(ptr.To(3), ptr.To(3)),
		},
		"lookup":   lookup(dot),
		"quantity": quantity(dot),
	}
}

func pod() *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "spacename",
			Name:      "eman",
		},
	}
}

func pdb() *policyv1.PodDisruptionBudget {
	minAvail := intstr.FromInt32(3)
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policyv1",
			Kind:       "PodDisruptionBudget",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvail,
		},
	}
}

func service() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{}, // Include an empty port to test the zero value of intstr.
			},
		},
	}
}

func lookup(dot *helmette.Dot) []any {
	svc, ok1 := helmette.Lookup[corev1.Service](dot, "namespace", "name")
	if !ok1 {
		panic(fmt.Sprintf("%T %q not found. Test setup should have created it?", corev1.Service{}, "name"))
	}

	sts, ok2 := helmette.Lookup[appsv1.StatefulSet](dot, "spacename", "eman")

	return []any{svc, ok1, sts, ok2}
}

func quantity(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	inputs := []string{
		"10",
		"100m", // 100 "milicores"
		"10G",
		"8Gi",
		"55Mi",
		"140k",
		// NB: Fractional values are intentionally left disabled as
		// resource.Quantity will rewrite these values to a normalized integral
		// form. This behavior is not required for correctness and therefore is
		// not (currently) being implemented.
		// "0.5",
		// "0.5Gi",
	}

	var quantities []resource.Quantity
	for _, in := range inputs {
		quantities = append(quantities, resource.MustParse(in))
	}

	if values.Quantity != nil {
		// NB: This is a bit of a hack. gotohelm's Unwrap will leave .Values
		// untouched as we expect it to a correct JSON representation of
		// .Values. This test receives a float64 as input for values.Quantity
		// which is a valid JSON representation as far as
		// resource.Quantity.UnmarshalJSON is concerned. However,
		// resource.Quantity.MarshalJSON always returns a string. To prevent
		// the test fixture from complaining about the difference, we copy the
		// quantity so gotohelm will actually transform the value to the go
		// equivalent.
		quantities = append(quantities, values.Quantity.DeepCopy())
	}

	var millis []int64
	var strs []string
	var value []int64
	for _, q := range quantities {
		millis = append(millis, q.MilliValue())
		strs = append(strs, q.String())
		value = append(value, q.Value())
	}

	return map[string]any{
		"MustParse": quantities,
		"Value":     value,
		"String":    strs,
	}
}
