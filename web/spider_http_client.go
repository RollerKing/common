package web

import (
	"encoding/json"
	"fmt"
	"github.com/qjpcpu/common/web/httpclient"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

var _client = NewClient()

func NewClient(timeout ...time.Duration) *HttpClient {
	var tm time.Duration
	if len(timeout) > 0 {
		tm = timeout[0]
	} else {
		tm = time.Second * 10
	}
	return &HttpClient{
		Client:   &http.Client{Timeout: tm},
		headlock: new(sync.RWMutex),
		Headers:  make(httpclient.Header),
	}
}

type HttpClient struct {
	Client   *http.Client
	headlock *sync.RWMutex
	Headers  httpclient.Header
	httpclient.Debugger
}

// config
func (client *HttpClient) AsDefault() {
	_client = client
}

func (client *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.Client.Do(req)
}

func (client *HttpClient) EnableCookie() {
	jar, _ := cookiejar.New(nil)
	client.Client.Jar = jar
}

func (c *HttpClient) SetHeaders(h httpclient.Header) error {
	c.headlock.Lock()
	defer c.headlock.Unlock()
	for k, v := range h {
		c.Headers[k] = v
	}
	return nil
}

func (c *HttpClient) Get(uri string) (res []byte, err error) {
	return httpclient.Get(c, uri)
}

func (c *HttpClient) PostForm(urlstr string, data httpclient.Form, extraHeader httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	for k, v := range extraHeader {
		if v != "" {
			hder[k] = v
		}
	}
	values := url.Values{}
	for k, v := range data {
		values.Set(k, fmt.Sprint(v))
	}
	return httpclient.HttpRequest(c, "POST", urlstr, hder, []byte(values.Encode()))
}

func (c *HttpClient) PostJSON(urlstr string, data interface{}, extraHeader httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/json; charset=UTF-8"
	for k, v := range extraHeader {
		if v != "" {
			hder[k] = v
		}
	}
	var payload []byte
	switch d := data.(type) {
	case string:
		payload = []byte(d)
	case []byte:
		payload = d
	case nil:
		// do nothing
	default:
		payload, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	return httpclient.HttpRequest(c, "POST", urlstr, hder, payload)
}

// defaults
func SetHeaders(h httpclient.Header) error {
	return _client.SetHeaders(h)
}

func Get(uri string) (res []byte, err error) {
	return _client.Get(uri)
}

func PostForm(urlstr string, data httpclient.Form, extraHeader httpclient.Header) (res []byte, err error) {
	return _client.PostForm(urlstr, data, extraHeader)
}

func PostJSON(urlstr string, data interface{}, extraHeader httpclient.Header) (res []byte, err error) {
	return _client.PostJSON(urlstr, data, extraHeader)
}
