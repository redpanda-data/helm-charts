// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package redpanda_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/ptr"
	"pgregory.net/rapid"

	"github.com/redpanda-data/redpanda-operator/charts/redpanda"
	"github.com/redpanda-data/redpanda-operator/pkg/rapidutil"
)

func TestStrategicMergePatch(t *testing.T) {
	cases := []struct {
		Name     string
		Override redpanda.PodTemplate
		Original corev1.PodTemplateSpec
		Expected corev1.PodTemplateSpec
	}{
		{
			Name: "zero-values",
			Expected: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{},
					Tolerations:  []corev1.Toleration{},
				},
			},
		},
		{
			Name: "override-only",
			Override: redpanda.PodTemplate{
				Labels:      map[string]string{"a": "b"},
				Annotations: map[string]string{"x": "y"},
				Spec: &applycorev1.PodSpecApplyConfiguration{
					SecurityContext: &applycorev1.PodSecurityContextApplyConfiguration{
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
					NodeSelector: map[string]string{},
					Tolerations:  []corev1.Toleration{},
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
					NodeSelector: map[string]string{},
					Tolerations:  []corev1.Toleration{},
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
				Spec: &applycorev1.PodSpecApplyConfiguration{
					SecurityContext: &applycorev1.PodSecurityContextApplyConfiguration{
						RunAsNonRoot: ptr.To(true),
						RunAsGroup:   ptr.To[int64](300),
					},
					Containers: []applycorev1.ContainerApplyConfiguration{
						{
							Name: ptr.To("insecure-container"),
							SecurityContext: &applycorev1.SecurityContextApplyConfiguration{
								Privileged: ptr.To(false),
								RunAsGroup: ptr.To[int64](6767),
							},
						},
						{
							Name: ptr.To("secure-container"),
							Env: []applycorev1.EnvVarApplyConfiguration{
								{Name: ptr.To("MY_VAR"), Value: ptr.To("Overridden")},
								{Name: ptr.To("EXTRA"), Value: ptr.To("extra")},
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
								{Name: "UNTOUCHED", Value: "default"},
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
					NodeSelector: map[string]string{},
					Tolerations:  []corev1.Toleration{},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    ptr.To[int64](100),
						RunAsGroup:   ptr.To[int64](300),
						RunAsNonRoot: ptr.To(true),
					},
					Containers: []corev1.Container{
						{
							Name: "secure-container",
							Env: []corev1.EnvVar{
								{Name: "MY_VAR", Value: "Overridden"},
								{Name: "UNTOUCHED", Value: "default"},
								{Name: "EXTRA", Value: "extra"},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(false),
							},
						},
						{
							Name: "insecure-container",
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(false),
								RunAsGroup: ptr.To[int64](6767),
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

func TestStrategicMergePatch_nopanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := rapid.MakeCustom[corev1.PodTemplateSpec](rapidutil.KubernetesTypes).Draw(t, "original")
		override := rapid.MakeCustom[redpanda.PodTemplate](rapidutil.KubernetesTypes).Draw(t, "override")

		// As we're doing a lot of reflect hackery and merging, use rapid to
		// ensure we've not missed any cases that could trigger a panic.
		assert.NotPanics(t, func() {
			redpanda.StrategicMergePatch(override, original)
		})
	})
}

func TestParseCLIArgs(t *testing.T) {
	testCases := []struct {
		In  []string
		Out map[string]string
	}{
		{
			In:  []string{},
			Out: map[string]string{},
		},
		{
			In: []string{
				"-foo=bar",
				"-baz 1",
				"--help", "topic",
			},
			Out: map[string]string{
				"-foo":   "bar",
				"-baz":   "1",
				"--help": "topic",
			},
		},
		{
			In: []string{
				"invalid",
				"-valid=bar",
				"--trailing spaces ",
				"--bare=",
				"ignored-perhaps-confusingly",
				"--final",
			},
			Out: map[string]string{
				"-valid":     "bar",
				"--trailing": "spaces ",
				"--bare":     "",
				"--final":    "",
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Out, redpanda.ParseCLIArgs(tc.In))
	}

	t.Run("NotPanics", rapid.MakeCheck(func(t *rapid.T) {
		// We could certainly be more inventive with
		// the inputs here but this is more of a
		// fuzz test than a property test.
		in := rapid.SliceOf(rapid.String()).Draw(t, "input")

		assert.NotPanics(t, func() {
			redpanda.ParseCLIArgs(in)
		})
	}))
}
