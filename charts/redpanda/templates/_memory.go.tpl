{{- /* Generated from "memory.go" */ -}}

{{- define "redpanda.RedpandaReserveMemory" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $rpMem_1 := $values.resources.memory.redpanda -}}
{{- if (and (ne (toJson $rpMem_1) "null") (ne (toJson $rpMem_1.reserveMemory) "null")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" ((div ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $rpMem_1.reserveMemory) ))) "r") | int64) ((mul (1024 | int) (1024 | int)))) | int64)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" ((add (((mulf (((get (fromJson (include "redpanda.ContainerMemory" (dict "a" (list $dot) ))) "r") | int64) | float64) 0.002) | float64) | int64) (200 | int64)) | int64)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaMemory" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $memory := ((0 | int64) | int64) -}}
{{- $containerMemory := ((get (fromJson (include "redpanda.ContainerMemory" (dict "a" (list $dot) ))) "r") | int64) -}}
{{- $rpMem_2 := $values.resources.memory.redpanda -}}
{{- if (and (ne (toJson $rpMem_2) "null") (ne (toJson $rpMem_2.memory) "null")) -}}
{{- $memory = ((div ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $rpMem_2.memory) ))) "r") | int64) ((mul (1024 | int) (1024 | int)))) | int64) -}}
{{- else -}}
{{- $memory = (((mulf ($containerMemory | float64) 0.8) | float64) | int64) -}}
{{- end -}}
{{- if (eq $memory (0 | int64)) -}}
{{- $_ := (fail "unable to get memory value redpanda-memory") -}}
{{- end -}}
{{- if (lt $memory (256 | int64)) -}}
{{- $_ := (fail (printf "%d is below the minimum value for Redpanda" $memory)) -}}
{{- end -}}
{{- if (gt ((add $memory ((get (fromJson (include "redpanda.RedpandaReserveMemory" (dict "a" (list $dot) ))) "r") | int64)) | int64) $containerMemory) -}}
{{- $_ := (fail (printf "Not enough container memory for Redpanda memory values where Redpanda: %d, reserve: %d, container: %d" $memory ((get (fromJson (include "redpanda.RedpandaReserveMemory" (dict "a" (list $dot) ))) "r") | int64) $containerMemory)) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $memory) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ContainerMemory" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne (toJson $values.resources.memory.container.min) "null") -}}
{{- $_is_returning = true -}}
{{- (dict "r" ((div ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $values.resources.memory.container.min) ))) "r") | int64) ((mul (1024 | int) (1024 | int)))) | int64)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" ((div ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $values.resources.memory.container.max) ))) "r") | int64) ((mul (1024 | int) (1024 | int)))) | int64)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

