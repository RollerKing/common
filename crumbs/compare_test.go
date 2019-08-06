package crumbs

import "testing"

func TestCompareStruct(t *testing.T) {
	cbFn := func(path string, l, r interface{}) {
		t.Fatalf("%s not equal l(%v) r(%v)", path, l, r)
	}
	type Struct2 struct {
		StringField       string
		StringPtrField    *string
		SliceElemPtrField []*int
	}
	type Address [32]byte
	type CStruct struct {
		StringField       string
		StringPtrField    *string
		IntField          int
		IntPtrField       *int
		SliceField        []int
		SliceElemPtrField []*int
		StructField       Struct2
		StructPtrField    *Struct2
		Addr              Address
	}
	str := "abcx"
	istr := "inner str"
	st2 := Struct2{
		StringPtrField: &istr,
	}
	i := 33
	ia, ib, ic := 100, 200, 300
	addr := Address{}
	addr[2] = 9
	obj := &CStruct{
		StringField:       "string field",
		StringPtrField:    &str,
		IntField:          1000,
		IntPtrField:       &i,
		SliceField:        []int{1, 2, 3},
		SliceElemPtrField: []*int{&ia, &ib, &ic},
		StructPtrField:    &st2,
		Addr:              addr,
	}
	str_1 := "abcx"
	istr_1 := "inner str"
	st2_1 := Struct2{
		StringPtrField: &istr_1,
	}
	i_1 := 33
	ia_1, ib_1, ic_1 := 100, 200, 300
	addr_1 := Address{}
	addr_1[2] = 9
	obj_1 := &CStruct{
		StringField:       "string field",
		StringPtrField:    &str_1,
		IntField:          1000,
		IntPtrField:       &i_1,
		SliceField:        []int{1, 2, 3},
		SliceElemPtrField: []*int{&ia_1, &ib_1, &ic_1},
		StructPtrField:    &st2_1,
		Addr:              addr_1,
	}
	DeepCompare(obj, obj_1, cbFn)
	new_i := 201
	obj_1.SliceElemPtrField[1] = &new_i
	DeepCompare(obj, obj_1, func(path string, l, r interface{}) {
		if path != ".SliceElemPtrField[1]" {
			t.Fatal("should not equal")
		}
	})
}
