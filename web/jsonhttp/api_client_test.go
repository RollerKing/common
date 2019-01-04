package jsonhttp

import (
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	res := make(map[string]interface{})
	if err := Get("http://httpbin.org/get", &res); err != nil {
		t.Fatal(err)
	}
	fmt.Println(Colored(res))
	res = make(map[string]interface{})
	if err := Post("http://httpbin.org/post", map[string]string{
		"name": "json",
	}, &res); err != nil {
		t.Fatal(err)
	}
	fmt.Println(Colored(res))
}
