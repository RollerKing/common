package fixture

import (
	"encoding/json"
	"reflect"
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
	Link string
	URL  *string
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
