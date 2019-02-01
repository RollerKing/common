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

// NewClient new client
func NewClient() *HttpClient {
	return &HttpClient{
		Client:         &http.Client{Timeout: 10 * time.Second},
		headlock:       new(sync.RWMutex),
		Headers:        make(httpclient.Header),
		IHTTPInspector: &httpclient.Debugger{},
	}
}

// HttpClient client
type HttpClient struct {
	Client   *http.Client
	headlock *sync.RWMutex
	Headers  httpclient.Header
	httpclient.IHTTPInspector
}

// Do do not invoke
func (client *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.Client.Do(req)
}

// EnableCookie use cookie
func (client *HttpClient) EnableCookie() *HttpClient {
	jar, _ := cookiejar.New(nil)
	client.Client.Jar = jar
	return client
}

// SetTimeout timeout
func (client *HttpClient) SetTimeout(tm time.Duration) *HttpClient {
	client.Client.Timeout = tm
	return client
}

// SetHeaders add headers
func (c *HttpClient) SetHeaders(h httpclient.Header) *HttpClient {
	c.headlock.Lock()
	defer c.headlock.Unlock()
	for k, v := range h {
		c.Headers[k] = v
	}
	return c
}

// Get get url
func (c *HttpClient) Get(uri string) (res []byte, err error) {
	return httpclient.Get(c, uri, c.Headers)
}

// GetWithParams with qs
func (c *HttpClient) GetWithParams(uri string, params interface{}) (res []byte, err error) {
	return httpclient.GetWithParams(c, uri, params, c.Headers)
}

// PostForm post form
func (c *HttpClient) PostForm(urlstr string, data httpclient.Form, extraHeader httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	for k, v := range extraHeader {
		if v != "" {
			hder[k] = v
		}
	}
	for k, v := range c.Headers {
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

// PostJSON post json
func (c *HttpClient) PostJSON(urlstr string, data interface{}, extraHeader httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/json; charset=UTF-8"
	for k, v := range extraHeader {
		if v != "" {
			hder[k] = v
		}
	}
	for k, v := range c.Headers {
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
