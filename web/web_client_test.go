package web

import (
	"encoding/json"
	"github.com/qjpcpu/common/web/httpclient"
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient()
	c.EnableCookie().SetDebug(true)
	res, err := c.Get("http://httpbin.org/get")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
	res, err = c.PostForm("http://httpbin.org/post", httpclient.Form{
		"a": "text",
		"b": 34,
	}, httpclient.Header{"Extract": "AAA"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
	res, err = c.PostJSON("http://httpbin.org/post", map[string]interface{}{
		"a": "text",
		"b": 34,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}

func TestResolve(t *testing.T) {
	c := NewClient()
	addr := "http://api.ipify.org/?format=json"
	res := struct {
		IP string `json:"ip"`
	}{}
	if err := c.GetResolver(&res, json.Unmarshal).Resolve(c.Get(addr)); err != nil {
		t.Fatal(err)
	}
	t.Log(res)
	if res.IP == "" {
		t.Fatal("resolv fail")
	}
}
