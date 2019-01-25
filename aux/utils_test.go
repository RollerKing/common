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
	if len(InteractStrings()) != 0 {
		t.Fatal("err")
	}
	if str := strings.Join(InteractStrings(arr), "."); str != "a.b" {
		t.Fatal(str)
	}
}

func TestOStrings(t *testing.T) {
	arr := []string{}
	if len(RemoveString(arr, "")) != 0 {
		t.Fatal("error")
	}
	res := strings.Join(RemoveString([]string{"a", "b", "c", "d", "b"}, "a"), ".")
	if res != "b.c.d.b" {
		t.Fatal(res)
	}
	res = strings.Join(RemoveString([]string{"a", "b", "c", "d", "b"}, "b"), ".")
	if res != "a.c.d" {
		t.Fatal(res)
	}
	res = strings.Join(RemoveString([]string{"a"}, "a"), ".")
	if res != "" {
		t.Fatal(res)
	}
	res = strings.Join(RemoveString([]string{"b", "a"}, "a"), ".")
	if res != "b" {
		t.Fatal(res)
	}
}
