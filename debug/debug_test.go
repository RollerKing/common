package debug

import (
	"testing"
)

type MyErr struct{}

func (MyErr) Error() string {
	return "err"
}

func TestShouldBeNil(t *testing.T) {
	var err error
	isPanic := func(err error) (y bool) {
		defer func() {
			if recover() != nil {
				y = true
			}
		}()
		ShouldBeNil(err,"but get %v",err)
		return
	}

	if isPanic(err) {
		t.Fatal("should not panic")
	}
	var m MyErr
	err = m
	if !isPanic(err) {
		t.Fatal("should panic")
	}
	var mm *MyErr
	err = mm
	if isPanic(err) {
		t.Fatal("should not panic")
	}
}
