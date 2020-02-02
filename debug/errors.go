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

// ShouldBeTrue would panic if codition is false
func ShouldBeTrue(condition bool) {
	if !condition {
		panic("should be true")
	}
}

// ShouldEqual would panic if not equal
func ShouldEqual(a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		panic(fmt.Sprintf("%v != %v", a, b))
	}
}

// AllowPanic swallow panic
func AllowPanic(fn func()) (isPanicOccur bool) {
	defer func() {
		if r := recover(); r != nil {
			isPanicOccur = true
		}
	}()
	fn()
	return
}
