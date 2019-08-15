package fixture

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestPickSimple(t *testing.T) {
	type SimpleStruct struct {
		String    string
		StringPtr *string
		NullPtr   *string
		Int       int
		Text      []string
	}
	s := &SimpleStruct{}
	if err := FillStruct(s); err != nil {
		t.Fatal(err)
	}
	s.NullPtr = nil

	res := PickValuesByLastNode(s, "String", "StringPtr", "Int", "Text[1]", "NullPtr")
	if res.MustGetString(".String") != s.String {
		t.Fatal("should get value")
	}
	if *(res.MustGetStringPtr(".StringPtr")) != *(s.StringPtr) {
		t.Fatal("should get value")
	}
	if e := res.MustGetStringPtr(".NullPtr"); e != nil {
		t.Fatal("should get value")
	}
	if res.MustGetInt(".Int") != int(s.Int) {
		t.Fatal("should get value")
	}
	if res.MustGetString(".Text[1]") != s.Text[1] {
		t.Fatal("should get value")
	}
}

func TestPickMap(t *testing.T) {
	type MapStruct struct {
		Map     map[string]string
		MapPtr  *map[string]int64
		MapPtr2 map[string]*int64
	}
	ms := &MapStruct{
		Map: map[string]string{
			"aaa": "v1",
			"bbb": "v2",
		},
		MapPtr: &map[string]int64{
			"ccc": 100,
			"ddd": 200,
		},
		MapPtr2: map[string]*int64{},
	}
	res := PickValuesByPath(ms, ".Map.aaa", ".MapPtr.ccc", ".MapPtr2.x")
	if res.MustGetString(".Map.aaa") != "v1" {
		t.Fatal("bad pick")
	}
	if res.MustGetInt64(".MapPtr.ccc") != 100 {
		t.Fatal("bad pick")
	}
	if _, err := res.Get(".MapPtr2.x"); err != ErrValueNotExist {
		t.Fatal("bad pick")
	}
}

func TestPickStruct(t *testing.T) {
	type A struct {
		Name   string
		B      *B
		M      map[string]*B
		MM     map[string]*int
		Nums   []*A
		Int    Integer
		Mobile string
		Tm     *time.Time
		ID     int
	}
	type B struct {
		Link   string
		URL    *string
		Phones []string
		Email  string
		Times  []*time.Time
	}
	a := &A{}
	if err := FillStruct(a, SetMaxLevel(3)); err != nil {
		t.Fatal("bad fill")
	}
	res := PickValues(a, func(string) bool { return true })
	if res.MustGetString(".Name") != a.Name {
		t.Fatal("bad pick")
	}
	if res.MustGetTimePtr(".B.Times[0]").Unix() != a.B.Times[0].Unix() {
		t.Fatal("bad pick")
	}
	if res.MustGetInt(".Int") != int(a.Int) {
		t.Fatal("bad pick")
	}
	if res.MustGetString(".B.Link") != a.B.Link {
		t.Fatal("bad pick")
	}
	if *(res.MustGetStringPtr(".B.URL")) != *(a.B.URL) {
		t.Fatal("bad pick")
	}
}

func TestPickerRecursive(t *testing.T) {
	type Node struct {
		ID       *string
		Children []*Node
	}
	n := &Node{}
	var id int64 = 1
	idMap := map[string]bool{}
	fillFn := func(path string, tp reflect.Type) (interface{}, bool) {
		if LastNodeOfPath(path) == "ID" {
			str := strconv.FormatInt(id, 10)
			idMap[str] = true
			id++
			return str, true
		}
		return nil, false
	}
	if err := FillStruct(n, WithSysPVFunc(fillFn), SetMaxLevel(3), SetMaxSliceLen(3)); err != nil {
		t.Fatal("fill err", err)
	}
	var pickIDs []string
	res := PickValuesByLastNode(n, "ID")
	keys := res.Paths()
	for _, k := range keys {
		if v := res.MustGetStringPtr(k); v != nil {
			pickIDs = append(pickIDs, *v)
		}
	}
	if len(pickIDs) != len(idMap) {
		t.Fatal("bad pick")
	}
	for _, id := range pickIDs {
		if !idMap[id] {
			t.Fatal("bad fill")
		}
	}
}
