package debug

import (
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/qjpcpu/qjson"
	"time"
)

var (
	colorFuncs = []func(a ...interface{}) string{
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgCyan).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgWhite, color.BgBlack).SprintFunc(),
		color.New(color.FgBlack, color.BgWhite).SprintFunc(),
	}
)

// Print with color
func Print(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Println(format)
		return
	}
	fmt.Printf(rewriteFormat(format, nil), colorArgs(rewriteArgsToString(format, args, false), nil)...)
}

// PrintWithTime print with time
func PrintWithTime(format string, args ...interface{}) {
	Print(timeStr(time.Now())+" "+format, args...)
}

// PrintJSONWithTime print with time
func PrintJSONWithTime(format string, args ...interface{}) {
	PrintJSON(timeStr(time.Now())+" "+format, args...)
}

// Debug print when debug on
func Debug(format string, args ...interface{}) {
	if IsDebug() {
		Print(format, args...)
	}
}

// PrintJSON complex value to json with color
func PrintJSON(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Println(format)
		return
	}
	withoutColor := make(map[int]bool)
	for i := range args {
		withoutColor[i] = isComplexValue(args[i])
	}
	fmt.Printf(rewriteFormat(format, nil), colorArgs(rewriteArgsToString(format, args, true), withoutColor)...)
}

// DebugJSON print when debug on
func DebugJSON(format string, args ...interface{}) {
	if IsDebug() {
		PrintJSON(format, args...)
	}
}

func rewriteArgsToString(format string, args []interface{}, complextToJSON bool) []interface{} {
	rewriteFormat(format, func(idx int, fmtToken string) {
		if idx >= len(args) {
			return
		}
		if complextToJSON && isComplexValue(args[idx]) {
			args[idx] = string(qjson.PrettyMarshal(args[idx]))
		} else {
			args[idx] = fmt.Sprintf(fmtToken, args[idx])
		}
	})
	return args
}

func rewriteFormat(format string, cb func(int, string)) string {
	if cb == nil {
		cb = func(int, string) {}
	}
	var idx int

	var newfmt []rune
	runes := []rune(format)
	for i := 0; i < len(runes); {
		/* skip double % */
		if runes[i] == '%' && i < len(runes)-1 && runes[i+1] == '%' {
			newfmt = append(newfmt, runes[i], runes[i+1])
			i += 2
			continue
		}
		/* find format token like %[^a-zA-Z] */
		if runes[i] == '%' {
			j := i + 1
			for ; j < len(runes); j++ {
				if (runes[j] >= 'A' && runes[j] <= 'Z') || (runes[j] >= 'a' && runes[j] <= 'z') {
					break
				}
			}
			cb(idx, string(runes[i:j+1]))
			idx++
			newfmt = append(newfmt, '%', 's')
			i = j + 1
			continue
		}
		newfmt = append(newfmt, runes[i])
		i++
	}
	/* always end with newline */
	if newfmt[len(newfmt)-1] != '\n' {
		newfmt = append(newfmt, '\n')
	}
	return string(newfmt)
}

func colorArgs(args []interface{}, withoutColor map[int]bool) []interface{} {
	ret := make([]interface{}, len(args))
	for i, v := range args {
		if withoutColor != nil && withoutColor[i] {
			ret[i] = args[i]
			continue
		}
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

func timeStr(tm time.Time) string {
	return tm.Format("15:04:05")
}
