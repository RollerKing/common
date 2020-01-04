package debug

import (
	"fmt"
	"reflect"
)

// ShouldSuccess would panic if err is not nil
func ShouldSuccess(err error) {
	if err == nil {
		return
	}
	v := reflect.ValueOf(err)
	if v.Kind() != reflect.Ptr || !v.IsNil() {
		panic(fmt.Sprintf("[%v]%v", v.Type(), err))
	}
}
