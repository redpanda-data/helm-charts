package redpanda_test

import (
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestStategicMergePatch(t *testing.T) {
	cases := []struct {
		Name     string
		Override redpanda.PodTemplate
		Original corev1.PodTemplateSpec
		Expected corev1.PodTemplateSpec
	}{
		{Name: "zero-values"},
		{
			Name: "override-only",
			Override: redpanda.PodTemplate{
				Labels:      map[string]string{"a": "b"},
				Annotations: map[string]string{"x": "y"},
				Spec: redpanda.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](100),
					},
				},
			},
			Original: corev1.PodTemplateSpec{},
			Expected: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"a": "b"},
					Annotations: map[string]string{"x": "y"},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](100),
					},
				},
			},
		},
		{
			Name:     "no-override",
			Override: redpanda.PodTemplate{},
			Original: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"a": "b"},
					Annotations: map[string]string{"x": "y"},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](100),
					},
				},
			},
			Expected: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"a": "b"},
					Annotations: map[string]string{"x": "y"},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](100),
					},
				},
			},
		},
		{
			Name: "merging",
			Override: redpanda.PodTemplate{
				Annotations: map[string]string{"a": "b"},
				Labels:      map[string]string{"x": "y"},
				Spec: redpanda.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						RunAsGroup:   ptr.To[int64](300),
					},
					Containers: []redpanda.Container{
						{
							Name: "insecure-container",
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(false),
							},
						},
						{
							Name: "secure-container",
							Env: []corev1.EnvVar{
								{Name: "MY_VAR", Value: "Overridden"},
							},
						},
					},
				},
			},
			Original: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"a": "b"},
					Annotations: map[string]string{"x": "y"},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  ptr.To[int64](100),
						RunAsGroup: ptr.To[int64](200),
					},
					Containers: []corev1.Container{
						{
							Name: "secure-container",
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(false),
							},
							Env: []corev1.EnvVar{
								{Name: "MY_VAR", Value: "default"},
							},
						},
						{
							Name: "insecure-container",
						},
					},
				},
			},
			Expected: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"a": "b", "x": "y"},
					Annotations: map[string]string{"a": "b", "x": "y"},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    ptr.To[int64](100),
						RunAsGroup:   ptr.To[int64](300),
						RunAsNonRoot: ptr.To(true),
					},
					Containers: []corev1.Container{
						{
							Name: "secure-container",
							Env: []corev1.EnvVar{
								{Name: "MY_VAR", Value: "default"},
								{Name: "MY_VAR", Value: "Overridden"},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(false),
							},
						},
						{
							Name: "insecure-container",
							Env:  []corev1.EnvVar{},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(false),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := redpanda.StrategicMergePatch(tc.Override, tc.Original)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
