package aux

import (
	"strings"
	"testing"
)

func TestAndStrings(t *testing.T) {
	arr := []string{"a", "b"}
	brr := []string{"e", "b"}
	crr := []string{"g"}
	res := MergeStrings(arr, brr, crr)
	if strings.Join(res, ".") != "a.b.e.g" {
		t.Fatal(res)
	}
	res = MergeStrings([]string{})
	if len(res) != 0 {
		t.Fatal("error")
	}

	res = SubstractStrings([]string{"a", "b", "c"}, []string{"b", "g"})
	if strings.Join(res, ".") != "a.c" {
		t.Fatal(res)
	}
	drr := []string{"b", "e"}
	if len(InteractStrings(arr, brr, crr, drr)) != 0 {
		t.Fatal("err")
	}
	res = InteractStrings(arr, brr, drr)
	if strings.Join(res, ".") != "b" {
		t.Fatal(res)
	}
}
