package web

import (
	"errors"
	"fmt"
	"github.com/qjpcpu/common/json"
	"github.com/qjpcpu/common/web/httpclient"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// NewClient new client
func NewClient() *HttpClient {
	return &HttpClient{
		Client:         &http.Client{Timeout: 5 * time.Second},
		IHTTPInspector: &httpclient.Debugger{},
	}
}

// HttpClient client
type HttpClient struct {
	Client *http.Client
	httpclient.IHTTPInspector
}

// ResponseResolver res resolver
type ResponseResolver struct {
	fn     httpclient.UnmarshalFunc
	resPtr interface{}
}

// Resolve response
func (rr *ResponseResolver) Resolve(data []byte, err error) error {
	if err != nil {
		return err
	}
	if rr.fn == nil || rr.resPtr == nil {
		return errors.New("bad http response resolver")
	}
	return rr.fn(data, rr.resPtr)
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
	if tm > time.Duration(0) {
		client.Client.Timeout = tm
	}
	return client
}

// Get get url
func (c *HttpClient) Get(uri string) (res []byte, err error) {
	return httpclient.Get(c, uri, nil)
}

// GetWithParams with qs
func (c *HttpClient) GetWithParams(uri string, params interface{}) (res []byte, err error) {
	return httpclient.GetWithParams(c, uri, params, nil)
}

// Post data
func (c *HttpClient) Post(urlstr string, data []byte, extraHeader httpclient.Header) (res []byte, err error) {
	return httpclient.HttpRequest(c, "POST", urlstr, extraHeader, data)
}

// PostForm post form
func (c *HttpClient) PostForm(urlstr string, data httpclient.Form, extraHeader httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/x-www-form-urlencoded"
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

// PostJSON post json
func (c *HttpClient) PostJSON(urlstr string, data interface{}, extraHeader httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/json"
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

// GetResolver get response resolver
func (c *HttpClient) GetResolver(resPtr interface{}, fn httpclient.UnmarshalFunc) *ResponseResolver {
	return &ResponseResolver{
		resPtr: resPtr,
		fn:     fn,
	}
}

// GetJSONResolver get json response resolver
func (c *HttpClient) GetJSONResolver(resPtr interface{}) *ResponseResolver {
	return c.GetResolver(resPtr, json.Unmarshal)
}
