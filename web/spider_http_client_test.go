package web

import (
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient()
	c.EnableCookie()
	if err := c.SetHeaders(map[string]string{
		"love":       "34",
		"user-agent": "safari",
	}); err != nil {
		t.Fatal(err)
	}
	res, err := c.Get("http://httpbin.org/get")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
	res, err = c.PostForm("http://httpbin.org/post", Form{
		"a": "text",
		"b": 34,
	}, Header{"Extract": "AAA"})
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
