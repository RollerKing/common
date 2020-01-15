package debug

import (
	"fmt"
	"reflect"
)

// ShouldBeNil would panic if err is not nil
func ShouldBeNil(err error) {
	if err == nil {
		return
	}
	v := reflect.ValueOf(err)
	if v.Kind() != reflect.Ptr || !v.IsNil() {
		panic(fmt.Sprintf("[%v]%v", v.Type(), err))
	}
}
