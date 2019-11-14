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
	if res.MustGet(".Int").Elem().Interface().(Integer) != a.Int {
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

func TestWalkStop(t *testing.T) {
	type X struct {
		Name string
		Age  int
	}
	x := &X{Name: "name", Age: 23}
	x1 := &X{}
	Walk(x, func(p string, tp reflect.Type, i ValuePtr) bool {
		if p == ".Name" {
			x1.Name = i.Elem().String()
			return false
		} else {
			return true
		}
	})
	if x1.Name != "name" {
		t.Fatal("bad walk")
	}
	if x1.Age != 0 {
		t.Fatal("bad walk")
	}
}

func TestWalkOnce(t *testing.T) {
	type X struct {
		Name *string
	}
	str := "ptr"
	cnt := 0
	x := &X{Name: &str}
	x1 := &X{}
	Walk(x, func(p string, tp reflect.Type, i ValuePtr) bool {
		if p == ".Name" {
			x1.Name = i.Elem().Interface().(*string)
			cnt++
		}
		return true
	})
	if *x1.Name != "ptr" {
		t.Fatal("bad walk")
	}
	if cnt != 1 {
		t.Fatal("bad walk")
	}
}

func TestWalkLeaf(t *testing.T) {
	type XI struct {
		Num int
	}
	type X struct {
		Name string
		Map  map[string]int
		List []XI
	}
	x := &X{
		Name: "name",
		Map:  map[string]int{"key": 12},
		List: []XI{
			{Num: 0},
			{Num: 1},
			{Num: 2},
		}}
	WalkLeaf(x, func(p string, tp reflect.Type, i ValuePtr) bool {
		switch p {
		case ".Name":
			if i.Elem().String() != "name" {
				t.Fatal("bad walk")
			}
			return true
		case ".Map.key":
			if i.Elem().Interface().(int) != 12 {
				t.Fatal("bad walk")
			}
			return true
		case ".List[0].Num":
			if i.Elem().Interface().(int) != 0 {
				t.Fatal("bad walk")
			}
			return true
		case ".List[1].Num":
			if i.Elem().Interface().(int) != 1 {
				t.Fatal("bad walk")
			}
			return true
		case ".List[2].Num":
			if i.Elem().Interface().(int) != 2 {
				t.Fatal("bad walk")
			}
			return true
		}
		t.Fatal("should not come here")
		return true
	})

}

func TestChangeMap(t *testing.T) {
	type Ss struct {
		Map map[string]string
	}

	s := &Ss{Map: make(map[string]string)}
	s.Map["key"] = "A"

	vals := PickValues(s, func(p string) bool {
		return p == ".Map.key"
	})

	vals.MustGet(".Map.key").Elem().SetString("B")

	if s.Map["key"] == "B" {
		t.Fatal("should not change map")
	}
}

func TestSetNil(t *testing.T) {
	type Object struct {
		StringPtr *string
	}
	s := "A"
	obj := &Object{StringPtr: &s}

	Walk(obj, func(p string, t reflect.Type, v ValuePtr) bool {
		if p == ".StringPtr" {
			v.Elem().Set(reflect.Zero(t))
		}
		return true
	})

	if obj.StringPtr != nil {
		t.Fatal("should be nil")
	}
}

func TestChangeSimpleValues(t *testing.T) {
	type Object struct {
		StringPtr *string
		String    string
		NumPtr    *int
		Num       int
	}
	s := "A"
	i := 1
	obj := &Object{StringPtr: &s, NumPtr: &i}

	Walk(obj, func(p string, t reflect.Type, v ValuePtr) bool {
		if p == ".StringPtr" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(&s2))
		} else if p == ".NumPtr" {
			i2 := 2
			v.Elem().Set(reflect.ValueOf(&i2))
		} else if p == ".Num" {
			v.Elem().SetInt(100)
		} else if p == ".String" {
			v.Elem().SetString("BTX")
		}
		return true
	})

	if *obj.StringPtr != "B" || *obj.NumPtr != 2 || obj.String != "BTX" || obj.Num != 100 {
		t.Fatal("bad change")
	}
}

func TestChangeStruct(t *testing.T) {
	type Object2 struct {
		StringPtr *string
		String    string
	}
	type Object struct {
		O    Object2
		OPtr *Object2
	}
	obj := &Object{OPtr: &Object2{}}

	Walk(obj, func(p string, t reflect.Type, v ValuePtr) bool {
		if p == ".O.StringPtr" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(&s2))
		} else if p == ".O.String" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(s2))
		} else if p == ".OPtr.StringPtr" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(&s2))
		} else if p == ".OPtr.String" {
			v.Elem().SetString("B")
		}
		return true
	})

	if *obj.O.StringPtr != "B" || obj.O.String != "B" || obj.OPtr.String != "B" || *obj.OPtr.StringPtr != "B" {
		t.Fatal("bad change")
	}

	Walk(obj, func(p string, t reflect.Type, v ValuePtr) bool {
		if p == ".O" {
			s := "A"
			o2 := Object2{String: "A", StringPtr: &s}
			v.Elem().Set(reflect.ValueOf(o2))
		} else if p == ".OPtr" {
			s := "A"
			o2 := &Object2{String: "A", StringPtr: &s}
			v.Elem().Set(reflect.ValueOf(o2))
		}
		return true
	})

	if *obj.O.StringPtr != "A" || obj.O.String != "A" || obj.OPtr.String != "A" || *obj.OPtr.StringPtr != "A" {
		t.Fatal("bad change")
	}
}

func TestChangeSlice(t *testing.T) {
	type Object2 struct {
		StringPtr *string
		String    string
	}
	type Object struct {
		List []Object2
		Ptr  []*Object2
	}
	strPtr := func(s string) *string { return &s }
	obj := &Object{}
	obj.Ptr = append(obj.Ptr, &Object2{String: "A", StringPtr: strPtr("A")})
	obj.List = append(obj.List, Object2{String: "A", StringPtr: strPtr("A")})

	Walk(obj, func(p string, t reflect.Type, v ValuePtr) bool {
		n := LastNodeOfPath(p)
		if n == "StringPtr" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(&s2))
		} else if n == "String" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(s2))
		}
		return true
	})

	if obj.Ptr[0].String != "B" || *obj.Ptr[0].StringPtr != "B" || obj.List[0].String != "B" || *obj.List[0].StringPtr != "B" {
		t.Fatal("bad change")
	}

	Walk(obj, func(p string, t reflect.Type, v ValuePtr) bool {
		if p == ".Ptr" {
			v.Elem().Set(reflect.Zero(t))
		}
		return true
	})

	if obj.Ptr != nil {
		t.Fatal("bad change")
	}
}

func TestCantChangeUnexportFields(t *testing.T) {
	type Object struct {
		str string
	}
	obj := &Object{}

	Walk(obj, func(p string, tp reflect.Type, v ValuePtr) bool {
		if p == ".str" {
			s2 := "B"
			v.Elem().Set(reflect.ValueOf(s2))
		}
		return true
	})
	if obj.str == "B" {
		t.Fatal("bad change")
	}
}

func TestVisitOnce(t *testing.T) {
	type Object2 struct {
		Str string
	}
	type Object struct {
		O *Object2
	}
	obj := &Object{O: &Object2{}}
	var count int
	Walk(obj, func(p string, tp reflect.Type, v ValuePtr) bool {
		if p == ".O" {
			count++
			if tp != reflect.TypeOf(new(Object2)) {
				t.Fatal("bad type")
			}
		}
		return true
	})
	if count != 1 {
		t.Fatal("bad visit")
	}
}

func TestChangeRootField(t *testing.T) {
	type Object struct {
		Str string
	}
	obj := &Object{}
	Walk(obj, func(p string, tp reflect.Type, v ValuePtr) bool {
		if tp == reflect.TypeOf(obj) {
			v.Elem().Elem().FieldByName("Str").SetString("A")
		}
		return true
	})
	if obj.Str != "A" {
		t.Fatal("bad change root")
	}
}
