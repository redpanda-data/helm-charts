package redpanda

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

// RedpandaReserveMemory will return the amount of memory that Redpanda process will not use from the provided value
// in `--memory` or from the internal Redpanda discovery process. It should be passed to the `--reserve-memory` argument
// of the Redpanda process, see RedpandaAdditionalStartFlags and rpk redpanda start documentation.
// https://docs.redpanda.com/current/reference/rpk/rpk-redpanda/rpk-redpanda-start/
func RedpandaReserveMemory(dot *helmette.Dot) int {
	values := helmette.Unwrap[Values](dot.Values)
	// This optional `redpanda` object allows you to specify the memory size for both the Redpanda
	// process and the underlying reserved memory used by Seastar.
	//
	// The amount of memory to allocate to a container is determined by the sum of three values:
	// 1. Redpanda (at least 2Gi per core, ~80% of the container's total memory)
	// 2. Seastar subsystem (200Mi * 0.2% of the container's total memory, 200Mi < x < 1Gi)
	// 3. Other container processes (whatever small amount remains)
	//if rpMem := values.Resources.Memory.Redpanda; rpMem != nil && rpMem.ReserveMemory != nil {
	if rpMem := values.Resources.Memory.Redpanda; rpMem != nil && rpMem.ReserveMemory != nil {
		// TODO currently string type cast will not work in go
		// error returned from the compiler :
		// invalid operation: rpMem.ReserveMemory (variable of type MemoryAmount) is not an interface
		if helmette.KindIs("string", rpMem.ReserveMemory) {
			return RedpandaMemoryToMi(*rpMem.ReserveMemory)
		}
		panic(fmt.Sprintf("Redpanda.ReserveMemory (%v) is not type of string", *rpMem.ReserveMemory))
	}

	// If Redpanda is omitted (default behavior), memory sizes are calculated automatically
	// based on 0.2% of container memory plus 200 Mi.
	return int(float64(ContainerMemory(dot))*0.002) + 200
}

// RedpandaMemory will return the amount of memory for Redpanda process. It should be passed to the
// `--memory` argument of the Redpanda process, see RedpandaAdditionalStartFlags and rpk redpanda start documentation.
// https://docs.redpanda.com/current/reference/rpk/rpk-redpanda/rpk-redpanda-start/
func RedpandaMemory(dot *helmette.Dot) int {
	values := helmette.Unwrap[Values](dot.Values)

	memory := 0
	containerMemory := ContainerMemory(dot)
	// This optional `redpanda` object allows you to specify the memory size for both the Redpanda
	// process and the underlying reserved memory used by Seastar.
	//
	// The amount of memory to allocate to a container is determined by the sum of three values:
	// 1. Redpanda (at least 2Gi per core, ~80% of the container's total memory)
	// 2. Seastar subsystem (200Mi * 0.2% of the container's total memory, 200Mi < x < 1Gi)
	// 3. Other container processes (whatever small amount remains)
	if rpMem := values.Resources.Memory.Redpanda; rpMem != nil && rpMem.Memory != nil {
		// TODO currently string type cast will not work in go
		// error returned from the compiler :
		// invalid operation: rpMem.Memory (variable of type MemoryAmount) is not an interface
		if helmette.KindIs("string", rpMem.Memory) {
			memory = int(RedpandaMemoryToMi(*rpMem.Memory))
		} else {
			panic(fmt.Sprintf("Redpanda.Memory (%v) is not type of string", *rpMem.ReserveMemory))
		}
	} else {
		//
		memory = int(float64(containerMemory) * 0.8)
	}

	if memory == 0 {
		panic("unable to get memory value redpanda-memory")
	}
	if memory < 256 {
		panic(fmt.Sprintf("%d is below the minimum value for Redpanda", memory))
	}

	if memory+int(RedpandaReserveMemory(dot)) > int(containerMemory) {
		panic(fmt.Sprintf("Not enough container memory for Redpanda memory values where Redpanda: %d, reserve: %d, container: %d", memory, RedpandaReserveMemory(dot), containerMemory))
	}

	return memory
}

// SIToBytes converts string representation of the bytes with SI suffixes to bytes
func SIToBytes(amount string) int {
	// Assert that the incoming amount is well formed.
	matched := helmette.RegexMatch(`^[0-9]+(\.[0-9]){0,1}(k|M|G|T|P|Ki|Mi|Gi|Ti|Pi)?$`, amount)
	if !matched {
		panic(fmt.Sprintf("amount (%s) does not match regex", amount))
	}

	unit := amount[len(amount)-1:]
	amount = amount[:len(amount)-1]

	// If we've got a _i suffix, pull out the full unit suffix.
	if unit == "i" {
		// TODO string + string not implemented.
		unit = fmt.Sprintf("%s%s", amount[len(amount)-1:], unit)
		amount = amount[:len(amount)-1]
	} else if helmette.RegexMatch(`\d`, unit) {
		// If unit is a number, we've gotten a raw byte amount.
		// TODO string + string not implemented.
		amount = fmt.Sprintf("%s%s", amount, unit)
		unit = ""
	}

	k := 1000
	m := k * k
	g := k * k * k
	t := k * k * k * k
	p := k * k * k * k * k

	ki := 1024
	mi := ki * ki
	gi := ki * ki * ki
	ti := ki * ki * ki * ki
	pi := ki * ki * ki * ki * ki

	amountFloat, err := helmette.Float64(amount)
	if err != nil {
		panic(fmt.Sprintf("SI to bytes conversion : %v", err))
	}

	// TODO(chrisseto): Really need to add in support for switch statements....
	if unit == "" {
		return int(amountFloat)
	} else if unit == "k" {
		return int(amountFloat * float64(k))
	} else if unit == "M" {
		return int(amountFloat * float64(m))
	} else if unit == "G" {
		return int(amountFloat * float64(g))
	} else if unit == "T" {
		return int(amountFloat * float64(t))
	} else if unit == "P" {
		return int(amountFloat * float64(p))
	} else if unit == "Ki" {
		return int(amountFloat * float64(ki))
	} else if unit == "Mi" {
		return int(amountFloat * float64(mi))
	} else if unit == "Gi" {
		return int(amountFloat * float64(gi))
	} else if unit == "Ti" {
		return int(amountFloat * float64(ti))
	} else if unit == "Pi" {
		return int(amountFloat * float64(pi))
	} else {
		panic(fmt.Sprintf("unknown unit: %q", unit))
	}
	//switch unit {
	//case "k":
	//	return int(amountFloat * float64(k))
	//case "M":
	//	return int(amountFloat * float64(m))
	//case "G":
	//	return int(amountFloat * float64(g))
	//case "T":
	//	return int(amountFloat * float64(t))
	//case "P":
	//	return int(amountFloat * float64(p))
	//case "Ki":
	//	return int(amountFloat * float64(ki))
	//case "Mi":
	//	return int(amountFloat * float64(mi))
	//case "Gi":
	//	return int(amountFloat * float64(gi))
	//case "Ti":
	//	return int(amountFloat * float64(ti))
	//case "Pi":
	//	return int(amountFloat * float64(pi))
	//default:
	//	panic(fmt.Sprintf("unknown unit: %q", unit))
	//}
}

func RedpandaMemoryToMi(amount MemoryAmount) int {
	return SIToBytes(string(amount)) / (1024 * 1024)
}

type MebiBytes = int

// Returns either the min or max container memory values as an integer value of MembiBytes
func ContainerMemory(dot *helmette.Dot) MebiBytes {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Resources.Memory.Container.Min != nil {
		return RedpandaMemoryToMi(*values.Resources.Memory.Container.Min)
	}

	return RedpandaMemoryToMi(values.Resources.Memory.Container.Max)
}
