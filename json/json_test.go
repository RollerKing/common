package json

import (
	"testing"
)

func TestDT(t *testing.T) {
	paths := []string{
		"a", "b", "a.c.d", "a.f.g",
	}
	dtree := createDelTree(paths)
	var nodes []string
	vf := func(n *delNode) bool {
		if n.val != "" {
			nodes = append(nodes, n.val)
		}
		return false
	}
	walkDelTreeBFS([]*delNode{dtree}, vf)
	t.Log(nodes)
}

func TestDel(t *testing.T) {
	js := `{"name":{"2m":3,"first":"Janet","arr":[1,2],"last":"Prichard","xx":{"yy":324,"tt":3}},"age":47,"weapon":"m4","tx":["s"]}`
	dels := []string{"name.first", "weapon", "name.xx.yy", "name.2m"}
	res := TrimJSON(js, dels...)
	if res != `{"name":{"arr":[1,2],"last":"Prichard","xx":{"tt":3}},"age":47,"tx":["s"]}` {
		t.Fatal(res)
	}
	js = ""
	res = TrimJSON(js, dels...)
	if res != "{}" {
		t.Fatal(res)
	}
	res = TrimJSON(js, "")
	if res != "{}" {
		t.Fatal(res)
	}
	js = `{"name":{"last":"Prichard","xx":{"yy":324,"tt":3}},"age":47,"100":"100",100:1}`
	res = TrimJSON(js, "100", "name.last", "name.xx")
	if res != `{"age":47}` {
		t.Fatal(res)
	}
	res = TrimJSON(js, "100", "name.last", "name.xx", "age")
	if res != `{}` {
		t.Fatal(res)
	}
}
