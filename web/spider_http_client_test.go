package web

import (
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient()
	c.EnableCookie()
	if err := c.SetHeaders("http://httpbin.org", map[string]string{
		"love":       "34",
		"user-agent": "safari",
	}); err != nil {
		t.Fatal(err)
	}
	c.InceptState()
	res, err := c.Get("http://httpbin.org/get")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
	res, err = c.PostForm("http://httpbin.org/post", map[string]interface{}{
		"a": "text",
		"b": 34,
	}, map[string]string{"Extract": "AAA"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
	res, err = c.PostJSON("http://httpbin.org/post", map[string]interface{}{
		"a": "text",
		"b": 34,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}
