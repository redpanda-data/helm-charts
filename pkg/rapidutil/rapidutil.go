// Package rapidutil contains utilities for working with the rapid property
// testing library. This primarily entails providing rapid Generators for
// Kubernetes types that would otherwise be considered invalid.
//
// For example, [intstr.IntOrString] may be created with a non existent
// [intstr.Type] without this package.
package rapidutil

import (
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"pgregory.net/rapid"
)

var (
	Quantity = rapid.Custom(func(t *rapid.T) *resource.Quantity {
		return resource.NewQuantity(rapid.Int64().Draw(t, "Quantity"), resource.DecimalSI)
	})

	Duration = rapid.Custom(func(t *rapid.T) metav1.Duration {
		dur := rapid.Int64().Draw(t, "Duration")
		return metav1.Duration{Duration: time.Duration(dur)}
	})

	IntOrString = rapid.Custom(func(t *rapid.T) intstr.IntOrString {
		if rapid.Bool().Draw(t, "intorstr") {
			return intstr.FromInt32(rapid.Int32().Draw(t, "FromInt32"))
		} else {
			return intstr.FromString(rapid.StringN(0, 10, 10).Draw(t, "FromString"))
		}
	})

	Probe = rapid.Custom(func(t *rapid.T) corev1.Probe {
		return corev1.Probe{
			InitialDelaySeconds: rapid.Int32Min(1).Draw(t, "InitialDelaySeconds"),
			FailureThreshold:    rapid.Int32Min(1).Draw(t, "FailureThreshold"),
			PeriodSeconds:       rapid.Int32Min(1).Draw(t, "PeriodSeconds"),
			TimeoutSeconds:      rapid.Int32Min(1).Draw(t, "TimeoutSeconds"),
			SuccessThreshold:    rapid.Int32Min(1).Draw(t, "SuccessThreshold"),
		}
	})

	KubernetesTypes = rapid.MakeConfig{
		Types: map[reflect.Type]*rapid.Generator[any]{
			reflect.TypeFor[int64]():              rapid.Int64Range(-99999, 99999).AsAny(),
			reflect.TypeFor[any]():                rapid.Just[any](nil), // Return nil for all untyped (any, interface{}) fields.
			reflect.TypeFor[*metav1.FieldsV1]():   rapid.Just[any](nil), // Return nil for K8s accounting fields.
			reflect.TypeFor[*resource.Quantity](): Quantity.AsAny(),
			reflect.TypeFor[metav1.Duration]():    Duration.AsAny(),
			reflect.TypeFor[intstr.IntOrString](): IntOrString.AsAny(),
			reflect.TypeFor[corev1.Probe]():       Probe.AsAny(),
		},
	}
)
