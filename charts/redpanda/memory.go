package redpanda

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
)

func RedPandaReserveMemory() {
}

func RedpandaMemory() {
}

func RedpandaMemoryFlags() []string {
	// Set in _configmap.tpl
	// - "--smp={{ include "redpanda-smp" . }}"
	// - "--memory={{ template "redpanda-memory" . }}M"
	// {{- if not .Values.config.node.developer_mode }}
	// - "--reserve-memory={{ template "redpanda-reserve-memory" . }}M"
	// {{- end }}

	return []string{}
}

func RedpandaMemoryToMi(amount string) int {
	return SIToBytes(amount) / (1024 * 1024)
}

func StatefulSetRedpandaContainerResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{}
	//	resources:
	//
	// {{- if hasKey .Values.resources.memory.container "min" }}
	//
	//	requests:
	//	  cpu: {{ .Values.resources.cpu.cores }}
	//	  memory: {{ .Values.resources.memory.container.min }}
	//
	// {{- end }}
	//
	//	limits:
	//	  cpu: {{ .Values.resources.cpu.cores }}
	//	  memory: {{ .Values.resources.memory.container.max }}
}

func SIToBytes(amount string) int {
	// Standardize units to be lowercased.
	amount = strings.ToLower(amount)
	// Assert that the incoming amount is well formed.
	helmette.MustRegexMatch(`^\d+(b|k|m|g|ki|mi|gi)?$`, amount)

	unit := string(amount[len(amount)-1])
	amount = amount[:len(amount)-1]

	// If we've got a _i suffix, pull out the full unit suffix.
	if unit == "i" {
		// TODO string + string not implemented.
		unit = fmt.Sprintf("%s%s", string(amount[len(amount)-1]), unit)
		amount = amount[:len(amount)-1]
	} else if helmette.RegexMatch(`\d`, unit) {
		// If unit is a number, we've gotten a raw byte amount.
		// TODO string + string not implemented.
		amount = fmt.Sprintf("%s%s", amount, unit)
		unit = "b"
	}

	k := 1000
	m := k * k
	g := k * k * k

	ki := 1024
	mi := ki * ki
	gi := ki * ki * ki

	amountInt := helmette.Atoi(amount)

	// TODO(chrisseto): Really need to add in support for switch statements....
	if unit == "b" {
		return amountInt
	} else if unit == "k" {
		return amountInt * k
	} else if unit == "m" {
		return amountInt * m
	} else if unit == "g" {
		return amountInt * g
	} else if unit == "ki" {
		return amountInt * ki
	} else if unit == "mi" {
		return amountInt * mi
	} else if unit == "gi" {
		return amountInt * gi
	} else {
		panic(fmt.Sprintf("unknown unit: %q", unit))
	}
	// switch unit {
	// case "b":
	// 	return amountInt
	// case "k":
	// 	return amountInt * k
	// case "m":
	// 	return amountInt * m
	// case "g":
	// 	return amountInt * g
	// case "ki":
	// 	return amountInt * ki
	// case "mi":
	// 	return amountInt * mi
	// case "gi":
	// 	return amountInt * gi
	// default:
	// 	panic(fmt.Sprintf("unknown unit: %q", unit))
	// }
}

func bytesToSI(amount int) string {
	return "TODO"
}

type MebiBytes = int

// Returns either the min or max container memory values as an integer value of MembiBytes
func ContainerMemory(dot *helmette.Dot) MebiBytes {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Resources.Memory.Container.Min != nil {
		asBytes := SIToBytes(string(*values.Resources.Memory.Container.Min))
		return asBytes / (1024 * 1024)
	}

	asBytes := SIToBytes(string(values.Resources.Memory.Container.Max))
	return asBytes / (1024 * 1024)
}
