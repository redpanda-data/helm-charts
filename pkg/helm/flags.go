package helm

import (
	"fmt"
	"reflect"
)

// ToFlags is a reflect based helper that translates a go struct with `flag`
// tags into a string slice of command line arguments.
// If flagsStruct is not a struct, ToFlags panics.
func ToFlags(flagsStruct any) []string {
	v := reflect.ValueOf(flagsStruct)
	t := v.Type()

	var flags []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		flag := field.Tag.Get("flag")
		if flag == "" || flag == "-" {
			continue
		}

		value := v.Field(i).Interface()

		if field.Type.Kind() == reflect.Bool && field.Name[:2] == "No" && flag[:3] != "no-" {
			value = !value.(bool)
		}

		if field.Type.Kind() == reflect.String && reflect.ValueOf(value).IsZero() {
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			for j := 0; j < v.Field(i).Len(); j++ {
				flags = append(flags, fmt.Sprintf("--%s=%v", flag, v.Field(i).Index(j)))
			}
			continue
		}

		flags = append(flags, fmt.Sprintf("--%s=%v", flag, value))
	}

	return flags
}
