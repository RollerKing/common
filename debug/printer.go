package debug

import (
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/qjpcpu/common/json"
)

var (
	colorFuncs = []func(a ...interface{}) string{
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
		color.New(color.FgCyan).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgWhite, color.BgBlack).SprintFunc(),
		color.New(color.FgBlack, color.BgWhite).SprintFunc(),
	}
)

// Print with color
func Print(format string, args ...interface{}) {
	fmt.Printf(rewriteFormat(format), colorArgs(args...)...)
}

// PrintJSON complex value to json with color
func PrintJSON(format string, args ...interface{}) {
	fmt.Printf(rewriteFormat(format), colorArgs(jsonArgs(args...)...)...)
}

func rewriteFormat(format string) string {
	var newfmt []rune
	runes := []rune(format)
	for i := 0; i < len(runes); {
		if runes[i] == '%' && i < len(runes)-1 && runes[i+1] == '%' {
			newfmt = append(newfmt, runes[i], runes[i+1])
			i += 2
			continue
		}
		if runes[i] == '%' {
			j := i + 1
			for ; j < len(runes); j++ {
				if (runes[j] >= 'A' && runes[j] <= 'Z') || (runes[j] >= 'a' && runes[j] <= 'z') {
					break
				}
			}
			newfmt = append(newfmt, '%', 's')
			i = j + 1
			continue
		}
		newfmt = append(newfmt, runes[i])
		i++
	}
	newfmt = append(newfmt, '\n')
	return string(newfmt)
}

func colorArgs(args ...interface{}) []interface{} {
	ret := make([]interface{}, len(args))
	for i, v := range args {
		ret[i] = colorFuncs[i%len(colorFuncs)](v)
	}
	return ret
}

func isComplexValue(v interface{}) bool {
	typ := reflect.TypeOf(v)
	if typ == nil {
		return false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Map, reflect.Struct, reflect.Slice:
		return true
	default:
		return false
	}
}

func jsonArgs(args ...interface{}) []interface{} {
	ret := make([]interface{}, len(args))
	for i, v := range args {
		if isComplexValue(v) {
			ret[i] = json.UnsafeMarshalString(v)
		} else {
			ret[i] = v
		}
	}
	return ret
}
