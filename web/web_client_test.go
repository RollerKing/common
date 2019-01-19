package web

import (
	"github.com/qjpcpu/common/web/httpclient"
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient()
	c.EnableCookie().SetDebug(true)
	c.SetHeaders(map[string]string{
		"love":       "34",
		"user-agent": "safari",
	})
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
