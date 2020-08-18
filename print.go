package factom

import (
	"fmt"
	"reflect"
	"strings"
)

const indent = "  "

func Print(v interface{}) string {
	return print(0, v)
}

func print(depth int, v interface{}) string {
	val := reflect.ValueOf(v)
	val = reflect.Indirect(val)

	typ := val.Type()

	sb := new(strings.Builder)
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		//f = reflect.Indirect(f)

		switch f.Kind() {
		case reflect.Struct:
			fmt.Fprintf(sb, "%s\n", typ.Field(i).Name)
			fmt.Fprintf(sb, "%s", print(depth+1, f.Interface()))
		case reflect.Slice:
			fmt.Fprintf(sb, "%s\n", typ.Field(i).Name)
			for i := 0; i < f.Len(); i++ {
				fmt.Fprintf(sb, "%s\n", print(depth+1, f.Index(i)))
			}
		default:
			fmt.Fprintf(sb, "%s = %v\n", typ.Field(i).Name, f)
		}
	}
	return sb.String()
}
