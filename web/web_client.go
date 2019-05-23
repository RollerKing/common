package web

import (
	"errors"
	"fmt"
	"github.com/qjpcpu/common/json"
	"github.com/qjpcpu/common/web/httpclient"
	"io"
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
		Client:     &http.Client{Timeout: 5 * time.Second},
		inspector:  httpclient.NewDebugger(),
		omit4xx5xx: false,
	}
}

// BeforeRequest modify req before request
type BeforeRequest func(*http.Request)

// HttpClient client
type HttpClient struct {
	Client       *http.Client
	inspector    *httpclient.Debugger
	globalHeader httpclient.Header
	beforeFunc   BeforeRequest
	omit4xx5xx   bool
}

// ResponseResolver res resolver
type ResponseResolver struct {
	fn         httpclient.UnmarshalFunc
	resPtr     interface{}
	omit4xx5xx bool
}

// Resolve response
func (rr *ResponseResolver) Resolve(data []byte, err error) error {
	if rr.fn == nil || rr.resPtr == nil {
		return errors.New("bad http response resolver")
	}
	if reflect.ValueOf(rr.resPtr).Kind() != reflect.Ptr {
		return errors.New("res obj must be pointer")
	}
	if err != nil {
		if rr.omit4xx5xx {
			// try to resolve http error content
			if he, ok := err.(*httpclient.HTTPError); ok {
				return rr.fn(he.Response, rr.resPtr)
			}
		}
		return err
	}

	return rr.fn(data, rr.resPtr)
}

// OmitHTTP4xx5xx omit http 4xx,5xx error
func (c *HttpClient) OmitHTTP4xx5xx(omit bool) *HttpClient {
	c.omit4xx5xx = omit
	return c
}

// SetBeforeFunc set before function
func (c *HttpClient) SetBeforeFunc(f BeforeRequest) *HttpClient {
	c.beforeFunc = f
	return c
}

// Do do not invoke
func (client *HttpClient) Do(req *http.Request) (*http.Response, error) {
	if f := client.beforeFunc; f != nil {
		f(req)
	}
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

// SetDebug set debug
func (c *HttpClient) SetDebug(on bool) *HttpClient {
	c.inspector.SetDebug(on)
	return c
}

// SetDebugWriter set debug writer
func (c *HttpClient) SetDebugWriter(w io.Writer) *HttpClient {
	c.inspector.SetWriter(w)
	return c
}

// IsDebugOn debug onoff
func (c *HttpClient) IsDebugOn() bool {
	return c.inspector.IsDebugOn()
}

// Inspect data
func (c *HttpClient) Inspect(data httpclient.TraceData) {
	c.inspector.Inspect(data)
}

// Get get url
func (c *HttpClient) Get(uri string, extraHeaders ...httpclient.Header) (res []byte, err error) {
	return httpclient.Get(c, uri, c.genHeaders(extraHeaders...))
}

// GetWithParams with qs
func (c *HttpClient) GetWithParams(uri string, params interface{}, extraHeaders ...httpclient.Header) (res []byte, err error) {
	return httpclient.GetWithParams(c, uri, params, c.genHeaders(extraHeaders...))
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

// HTTP any request
func (c *HttpClient) HTTP(method, urlstr string, data []byte, extraHeaders ...httpclient.Header) (res []byte, err error) {
	return httpclient.HttpRequest(c, method, urlstr, c.genHeaders(extraHeaders...), data)
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
		resPtr:     resPtr,
		fn:         fn,
		omit4xx5xx: c.omit4xx5xx,
	}
}

// GetJSONResolver get json response resolver
func (c *HttpClient) GetJSONResolver(resPtr interface{}) *ResponseResolver {
	return c.GetResolver(resPtr, json.Unmarshal)
}
