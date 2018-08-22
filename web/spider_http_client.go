package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
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
		Headers:  make(http.Header),
	}
}

type HttpClient struct {
	Client   *http.Client
	headlock *sync.RWMutex
	Headers  http.Header
}

type Header map[string]string
type Form map[string]interface{}

// config
func (client *HttpClient) AsDefault() {
	_client = client
}

func (client *HttpClient) EnableCookie() {
	jar, _ := cookiejar.New(nil)
	client.Client.Jar = jar
}

func (c *HttpClient) SetHeaders(h Header) error {
	c.headlock.Lock()
	defer c.headlock.Unlock()
	for k, v := range h {
		c.Headers.Set(k, v)
	}
	return nil
}

func (c *HttpClient) fillReqHeader(req *http.Request) {
	c.headlock.RLock()
	defer c.headlock.RUnlock()
	for name := range c.Headers {
		if v := c.Headers.Get(name); v != "" {
			req.Header.Set(name, v)
		}
	}
}

func (c *HttpClient) Get(uri string) (res []byte, err error) {
	return c.HttpRequest("GET", uri, nil, nil)
}

func (c *HttpClient) PostForm(urlstr string, data Form, extraHeader Header) (res []byte, err error) {
	hder := make(Header)
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
	return c.HttpRequest("POST", urlstr, hder, []byte(values.Encode()))
}

func (c *HttpClient) PostJSON(urlstr string, data interface{}, extraHeader Header) (res []byte, err error) {
	hder := make(Header)
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
	return c.HttpRequest("POST", urlstr, hder, payload)
}

func (c *HttpClient) HttpRequest(method, urlstr string, headers Header, bodyData []byte) (res []byte, err error) {
	var req *http.Request
	var body_data io.Reader
	method = strings.ToUpper(method)
	if !strings.HasPrefix(urlstr, "http://") && !strings.HasPrefix(urlstr, "https://") {
		urlstr = "http://" + urlstr
	}
	if bodyData != nil && len(bodyData) > 0 {
		body_data = bytes.NewBuffer(bodyData)
	}
	req, err = http.NewRequest(method, urlstr, body_data)
	if err != nil {
		return res, err
	}
	c.fillReqHeader(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rs, err := c.Client.Do(req)
	if err != nil {
		return res, err
	}
	defer rs.Body.Close()
	res, err = ioutil.ReadAll(rs.Body)
	if err != nil {
		return res, err
	}
	return res, nil
}

// defaults
func SetHeaders(h Header) error {
	return _client.SetHeaders(h)
}

func Get(uri string) (res []byte, err error) {
	return _client.Get(uri)
}

func PostForm(urlstr string, data Form, extraHeader Header) (res []byte, err error) {
	return _client.PostForm(urlstr, data, extraHeader)
}

func PostJSON(urlstr string, data interface{}, extraHeader Header) (res []byte, err error) {
	return _client.PostJSON(urlstr, data, extraHeader)
}
