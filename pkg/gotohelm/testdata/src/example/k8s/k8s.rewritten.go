//go:build rewrites
package k8s

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

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
		"lookup": lookup(dot),
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
	tmp_tuple_1 := helmette.Compact2(helmette.Lookup[corev1.Service](dot, "namespace", "name"))
	ok1 := tmp_tuple_1.T2
	svc := tmp_tuple_1.T1
	if !ok1 {
		panic(fmt.Sprintf("%T %q not found. Test setup should have created it?", corev1.Service{}, "name"))
	}
	tmp_tuple_2 := helmette.Compact2(helmette.Lookup[appsv1.StatefulSet](dot, "spacename", "eman"))
	ok2 := tmp_tuple_2.T2
	sts := tmp_tuple_2.T1

	return []any{svc, ok1, sts, ok2}
}
