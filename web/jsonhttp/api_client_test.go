package jsonhttp

import (
	"testing"
)

func TestHttp(t *testing.T) {
	client := NewClient()
	client.SetDebug(true)
	res := make(map[string]interface{})
	if err := client.Get("http://httpbin.org/get", &res); err != nil {
		t.Fatal(err)
	}
	params := struct {
		Int    int
		Text   string
		IDList []string
		ShopID string
	}{
		Int:    23,
		Text:   "text",
		ShopID: "32",
		IDList: []string{"id1", "id2"},
	}
	if err := client.GetWithParams("http://httpbin.org/get", params, &res); err != nil {
		t.Fatal(err)
	}
	res = make(map[string]interface{})
	if err := client.Post("http://httpbin.org/post", map[string]string{
		"name": "json",
	}, &res); err != nil {
		t.Fatal(err)
	}
}
