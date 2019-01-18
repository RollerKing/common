package json

import (
	"fmt"
	"testing"
)

func TestDT(t *testing.T) {
	paths := []string{
		"a", "b", "a.c.d", "a.f.g",
	}
	dtree := createDelTree(paths)
	vf := func(n *delNode) bool {
		fmt.Println(n.val)
		return false
	}
	walkDelTreeBFS([]*delNode{dtree}, vf)
}

func TestDel(t *testing.T) {
	js := `{"name":{"2m":3,"first":"Janet","arr":[1,2],"last":"Prichard","xx":{"yy":324,"tt":3}},"age":47,"weapon":"m4","tx":["s"]}`
	fmt.Println(js)
	dels := []string{"name.first", "weapon", "name.xx.yy", "name.2m"}
	res := TrimJSON(js, dels...)
	fmt.Println(res)
	js = ""
	fmt.Println(TrimJSON(js, dels...))
	fmt.Println(TrimJSON(js, ""))
	js = `{"name":{"last":"Prichard","xx":{"yy":324,"tt":3}},"age":47,"100":"100",100:1}`
	fmt.Println(TrimJSON(js, "100", "name.last", "name.xx"))
	fmt.Println(TrimJSON(js, "100", "name.last", "name.xx", "age"))
}
