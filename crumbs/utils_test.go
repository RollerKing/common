package crumbs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAndStrings(t *testing.T) {
	arr := []string{"a", "b"}
	brr := []string{"e", "b"}
	crr := []string{"g"}
	res := UnionStrings(arr, brr, crr)
	if strings.Join(res, ".") != "a.b.e.g" {
		t.Fatal(res)
	}
	res = UnionStrings([]string{})
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

func TestStruct2Map(t *testing.T) {
	i := 12
	obj := struct {
		Time   time.Time `json:"time"`
		Text   string    `json:"txt,omitempty"`
		Num    int       `json:"num"`
		NumPtr *int      `json:"num_ptr,omitempty"`
		Objs   []struct {
			A string
			B int `json:"b"`
		}
	}{
		NumPtr: &i,
		Objs: []struct {
			A string
			B int `json:"b"`
		}{
			{
				A: "aaa",
				B: 23,
			},
		},
	}
	m := Struct2Map(&obj)
	data, _ := json.Marshal(m)
	t.Log(string(data))
}
