//go:build rewrites
package k8s

import (
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func K8s() map[string]any {
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
