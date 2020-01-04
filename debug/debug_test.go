package debug

import (
	"testing"
)

type MyErr struct{}

func (MyErr) Error() string {
	return "err"
}

func TestShouldSuccess(t *testing.T) {
	var err error
	isPanic := func(err error) (y bool) {
		defer func() {
			if recover() != nil {
				y = true
			}
		}()
		ShouldSuccess(err)
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
