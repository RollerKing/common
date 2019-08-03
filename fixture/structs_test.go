package fixture

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

type Integer int
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
	Text   []*C
}

type B struct {
	Link   string
	URL    *string
	Phones []string
	Email  string
	C      *C
	Times  []*time.Time
}
type C struct {
	Link       string
	URL        *string
	MapInteger map[string]*Integer
}

func TestFillStruct(t *testing.T) {
	obj := &A{}
	now := time.Now()
	fn := func(path string, tp reflect.Type) (interface{}, bool) {
		t.Logf("path=%s type=%s", path, tp)
		switch path {
		case "B.Link":
			return "http://www.github.com", true
		case "B.Times.[0]":
			return &now, true
		case "B.C.Link":
			return "should drop this value", true
		case "Int":
			return 1024, true
		case "B.URL":
			return nil, true
		}
		return nil, false
	}
	if err := FillStruct(obj, SetMaxLevel(2), SetMaxMapLen(1), SetMaxSliceLen(1), WithSysPVFunc(fn)); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(obj, "", "   ")
	t.Log(string(data))
	if obj.B.Times[0].UnixNano() != now.UnixNano() {
		t.Fatal("fill bad value")
	}
	if obj.B.Link != "http://www.github.com" {
		t.Fatal("fill bad value")
	}
	if obj.B.C.Link != "" {
		t.Fatal("fill bad value")
	}
	if int(obj.Int) != 1024 {
		t.Fatal("fill bad value")
	}
}

type node struct {
	Left  *node
	Right *node
	Val   int
}

func TestMaxSliceLen(t *testing.T) {
	b := &B{}
	FillStruct(b, SetMaxSliceLen(2), SetMaxMapLen(3))
	if len(b.Phones) != 2 || len(b.Times) != 2 || len(b.C.MapInteger) != 3 {
		t.Fatal("fill bad value")
	}
}

type MMap map[string]*MMap

func TestMaxLevel(t *testing.T) {
	root := &node{}
	lvl := 20
	FillStruct(root, SetMaxLevel(lvl), SetPathToValueFunc(func(p string, tp reflect.Type) (interface{}, bool) {
		if strings.HasSuffix(p, "Val") {
			return 5, true
		}
		return nil, false
	}))
	cur := root
	for i := 0; i < lvl; i++ {
		if i < lvl-1 {
			if cur.Val != 5 {
				t.Fatal("fill bad value")
			}
		}
		cur = cur.Left
	}
	if cur.Val != 0 {
		t.Fatal("fill bad value")
	}

	aa := &A{}
	FillStruct(aa, SetMaxLevel(1))
	if aa.B.Link != "" {
		t.Fatal("fill bad value")
	}
}

func TestFillNonStruct(t *testing.T) {
	var strs []string
	FillStruct(&strs)
	if len(strs) != 3 || strs[0] == "" {
		t.Fatal("fill bad value")
	}
	m := make(map[string]int)
	FillStruct(&m)
	if len(m) == 0 {
		t.Fatal("fill bad value")
	}

	type Address [5]byte
	var array Address
	if err := FillStruct(&array); err != nil {
		t.Fatal(err)
	}

	obj2 := struct {
		Addr  Address
		Addr2 *Address
	}{}
	if err := FillStruct(&obj2); err != nil {
		t.Fatal(err)
	}
}
