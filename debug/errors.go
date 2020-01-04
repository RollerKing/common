package debug

import (
	"reflect"
)

func ShouldSuccess(err error) {
	if err == nil {
		return
	}
	v := reflect.ValueOf(err)
	if !v.IsNil() {
		panic(err)
	}
}
