package web

import (
	"errors"
	"fmt"
	"github.com/qjpcpu/common/json"
	"github.com/qjpcpu/common/web/httpclient"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"reflect"
	"time"
)

// NewClient new client
func NewClient() *HttpClient {
	return &HttpClient{
		Client:         &http.Client{Timeout: 5 * time.Second},
		IHTTPInspector: httpclient.NewDebugger(),
	}
}

// HttpClient client
type HttpClient struct {
	Client *http.Client
	httpclient.IHTTPInspector
	globalHeader httpclient.Header
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
	if reflect.ValueOf(rr.resPtr).Kind() != reflect.Ptr {
		return errors.New("res obj must be pointer")
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

// SetGlobalHeader set global headers, should be set before any request happens(cas unsafe map)
func (client *HttpClient) SetGlobalHeader(name, val string) *HttpClient {
	if name == "" {
		return client
	}
	name = textproto.CanonicalMIMEHeaderKey(name)
	if client.globalHeader == nil {
		client.globalHeader = make(httpclient.Header)
	}
	if val == "" {
		if _, ok := client.globalHeader[name]; ok {
			delete(client.globalHeader, name)
		}
		return client
	}
	client.globalHeader[name] = val
	return client
}

// Get get url
func (c *HttpClient) Get(uri string) (res []byte, err error) {
	return httpclient.Get(c, uri, c.genHeaders())
}

// GetWithParams with qs
func (c *HttpClient) GetWithParams(uri string, params interface{}) (res []byte, err error) {
	return httpclient.GetWithParams(c, uri, params, c.genHeaders())
}

// Post data
func (c *HttpClient) Post(urlstr string, data []byte, extraHeaders ...httpclient.Header) (res []byte, err error) {
	return httpclient.HttpRequest(c, "POST", urlstr, c.genHeaders(extraHeaders...), data)
}

// PostForm post form
func (c *HttpClient) PostForm(urlstr string, data httpclient.Form, extraHeaders ...httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/x-www-form-urlencoded"
	for _, extraHeader := range extraHeaders {
		for k, v := range extraHeader {
			hder[textproto.CanonicalMIMEHeaderKey(k)] = v
		}
	}
	values := url.Values{}
	for k, v := range data {
		values.Set(k, fmt.Sprint(v))
	}
	return httpclient.HttpRequest(c, "POST", urlstr, c.genHeaders(hder), []byte(values.Encode()))
}

// PostJSON post json
func (c *HttpClient) PostJSON(urlstr string, data interface{}, extraHeaders ...httpclient.Header) (res []byte, err error) {
	hder := make(httpclient.Header)
	hder["Content-Type"] = "application/json"
	for _, extraHeader := range extraHeaders {
		for k, v := range extraHeader {
			hder[textproto.CanonicalMIMEHeaderKey(k)] = v
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
	return httpclient.HttpRequest(c, "POST", urlstr, c.genHeaders(hder), payload)
}

func (c *HttpClient) genHeaders(extraHeaders ...httpclient.Header) httpclient.Header {
	if len(c.globalHeader) == 0 && len(extraHeaders) == 0 {
		return nil
	}
	hder := make(httpclient.Header)
	for key, val := range c.globalHeader {
		key = textproto.CanonicalMIMEHeaderKey(key)
		if val != "" {
			hder[key] = val
		}
	}
	for _, sub := range extraHeaders {
		for key, val := range sub {
			key = textproto.CanonicalMIMEHeaderKey(key)
			if val == "" {
				if _, ok := hder[key]; ok {
					delete(hder, key)
				}
			} else {
				hder[key] = val
			}
		}
	}
	return hder
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
