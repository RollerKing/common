package crumbs

import (
	"testing"
)

func testCloneObj(toPtr, fromPtr interface{}, t *testing.T, desc string) {
	err := Clone(toPtr, fromPtr)
	if err != nil {
		t.Fatalf("[%s]clone fail:%v", desc, err)
	}
	cbFn := func(path string, l, r interface{}) {
		t.Fatalf("[%s]clone fail %s from(%v) != to(%v)", desc, path, l, r)
	}
	DeepCompare(fromPtr, toPtr, cbFn)
}

func TestCloneSimple(t *testing.T) {
	ifrom := 23
	var ito int
	desc := "SimpleClone"
	testCloneObj(&ito, &ifrom, t, desc)

	strFrom := "gogogo"
	var strTo string
	testCloneObj(&strTo, &strFrom, t, desc)
}

func TestCloneSlice(t *testing.T) {
	desc := "SliceClone"
	text := []string{"a", "b", "c"}
	var text2 []string
	testCloneObj(&text2, &text, t, desc)

	str1, str2, str3 := "aaa", "bbb", "ccc"
	textPtr := []*string{&str1, &str2, &str3}
	var toPtr []*string
	testCloneObj(&toPtr, &textPtr, t, desc)

	*(textPtr[2]) = "new text"
	if *(toPtr[2]) != "ccc" {
		t.Fatal("cloned slice should not be changed")
	}

	array := [3]byte{5, 9, 2}
	var toArray [3]byte
	testCloneObj(&toArray, &array, t, desc)
}

func TestCloneMap(t *testing.T) {
	m1 := map[string]int{
		"key1": 1,
		"key2": 2,
	}
	m2 := make(map[string]int)
	desc := "MapClone"
	testCloneObj(&m2, &m1, t, desc)
}

func TestCloneStruct(t *testing.T) {
	desc := "StructClone"
	type MyStruct struct {
		FString    string
		FStringPtr *string
		FSlice     []int
	}
	str := "string ptr"
	ms1 := MyStruct{
		FString:    "aaa",
		FStringPtr: &str,
		FSlice:     []int{3, 4, 5},
	}
	var ms2 MyStruct
	testCloneObj(&ms2, &ms1, t, desc)

	ms1.FSlice[1] = 100
	if ms2.FSlice[1] == 100 {
		t.Fatal("cloned obj should not be changed")
	}
	str2 := "new string"
	ms1.FStringPtr = &str2
	if *ms2.FStringPtr == str2 {
		t.Fatal("cloned obj should not be changed")
	}
}

func TestCloneBadParam(t *testing.T) {
	var str string
	if err := Clone(nil, nil); err != nil {
		t.Fatal("both nil is ok")
	}
	if err := Clone(nil, &str); err == nil {
		t.Fatal("should not clone any nil")
	}
	if err := Clone(&str, nil); err == nil {
		t.Fatal("should not clone any nil")
	}
	var i int
	if err := Clone(&str, &i); err == nil {
		t.Fatal("should not clone within different type")
	}
	if err := Clone(&str, &str); err == nil {
		t.Fatal("should not clone same object")
	}
}

func TestCloneNil(t *testing.T) {
	desc := "NilClone"
	type NilStruct struct {
		Map      map[int]int
		Slice    []string
		Slice2   []int
		Anything interface{}
	}
	n1 := NilStruct{Slice2: make([]int, 2)}
	var n2 NilStruct
	testCloneObj(&n2, &n1, t, desc)
	if n2.Slice != nil || n2.Map != nil {
		t.Fatal("should be nil")
	}
	if len(n2.Slice2) != 2 {
		t.Fatal("should be len 2")
	}
}

func TestCloneInterface(t *testing.T) {
	desc := "InterfaceClone"
	type NilStruct struct {
		Anything interface{}
	}
	s := "string"
	n1 := NilStruct{Anything: s}
	var n2 NilStruct
	testCloneObj(&n2, &n1, t, desc)
}
