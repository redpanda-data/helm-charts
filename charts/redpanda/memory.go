// +gotohelm:filename=_memory.go.tpl
package redpanda

import (
	"fmt"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

type MebiBytes = int64

// RedpandaReserveMemory will return the amount of memory that Redpanda process will not use from the provided value
// in `--memory` or from the internal Redpanda discovery process. It should be passed to the `--reserve-memory` argument
// of the Redpanda process, see RedpandaAdditionalStartFlags and rpk redpanda start documentation.
// https://docs.redpanda.com/current/reference/rpk/rpk-redpanda/rpk-redpanda-start/
func RedpandaReserveMemory(dot *helmette.Dot) int64 {
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
		return rpMem.ReserveMemory.Value() / (1024 * 1024)
	}

	// If Redpanda is omitted (default behavior), memory sizes are calculated automatically
	// based on 0.2% of container memory plus 200 Mi.
	return int64(float64(ContainerMemory(dot))*0.002) + 200
}

// RedpandaMemory will return the amount of memory for Redpanda process. It should be passed to the
// `--memory` argument of the Redpanda process, see RedpandaAdditionalStartFlags and rpk redpanda start documentation.
// https://docs.redpanda.com/current/reference/rpk/rpk-redpanda/rpk-redpanda-start/
func RedpandaMemory(dot *helmette.Dot) int64 {
	values := helmette.Unwrap[Values](dot.Values)

	memory := int64(0)
	containerMemory := ContainerMemory(dot)
	// This optional `redpanda` object allows you to specify the memory size for both the Redpanda
	// process and the underlying reserved memory used by Seastar.
	//
	// The amount of memory to allocate to a container is determined by the sum of three values:
	// 1. Redpanda (at least 2Gi per core, ~80% of the container's total memory)
	// 2. Seastar subsystem (200Mi * 0.2% of the container's total memory, 200Mi < x < 1Gi)
	// 3. Other container processes (whatever small amount remains)
	if rpMem := values.Resources.Memory.Redpanda; rpMem != nil && rpMem.Memory != nil {
		memory = rpMem.Memory.Value() / (1024 * 1024)
	} else {
		//
		memory = int64(float64(containerMemory) * 0.8)
	}

	if memory == 0 {
		panic("unable to get memory value redpanda-memory")
	}

	if memory < 256 {
		panic(fmt.Sprintf("%d is below the minimum value for Redpanda", memory))
	}

	if memory+RedpandaReserveMemory(dot) > containerMemory {
		panic(fmt.Sprintf("Not enough container memory for Redpanda memory values where Redpanda: %d, reserve: %d, container: %d", memory, RedpandaReserveMemory(dot), containerMemory))
	}

	return memory
}

// Returns either the min or max container memory values as an integer value of MembiBytes
func ContainerMemory(dot *helmette.Dot) MebiBytes {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Resources.Memory.Container.Min != nil {
		return values.Resources.Memory.Container.Min.Value() / (1024 * 1024)
	}

	return values.Resources.Memory.Container.Max.Value() / (1024 * 1024)
}
