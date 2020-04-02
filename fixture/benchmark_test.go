package fixture

import (
	"reflect"
	"testing"
)

func BenchmarkWalk(b *testing.B) {
	obj := makeComplexObject()
	for i := 0; i < b.N; i++ {
		walkObject(obj)
	}
}

type SubObject struct {
	Num int
}

type MainObject struct {
	Name string
	Map  map[string]int
	List []SubObject
}

func makeComplexObject() interface{} {
	obj := &MainObject{}
	FillStruct(&obj)
	return obj
}

func walkObject(obj interface{}) {
	Walk(obj, func(p string, tp reflect.Type, i ValuePtr) bool {
		return true
	})
}
