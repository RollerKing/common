package jsonhttp

import (
	"errors"
	sysjson "github.com/qjpcpu/common/json"
	"github.com/qjpcpu/common/web/httpclient"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"time"
)

// JSONClient would auto parse response as JSON
type JSONClient struct {
	*http.Client
	httpclient.IHTTPInspector
}

// NewClient create new json client
func NewClient() *JSONClient {
	jc := &JSONClient{
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		IHTTPInspector: &httpclient.Debugger{},
	}
	jc.SetDebug(false)
	return jc
}

// EnableCookie would keep cookie info
func (jc *JSONClient) EnableCookie(cookie bool) *JSONClient {
	if cookie {
		jar, _ := cookiejar.New(nil)
		jc.Client.Jar = jar
	} else {
		jc.Client.Jar = nil
	}
	return jc
}

// SetTimeout set timeout
func (jc *JSONClient) SetTimeout(tm time.Duration) *JSONClient {
	jc.Client.Timeout = tm
	return jc
}

// Do http request should not invoke by user
func (js *JSONClient) Do(req *http.Request) (*http.Response, error) {
	return js.Client.Do(req)
}

// GetWithParams get url by encode params as query string
func (jc *JSONClient) GetWithParams(urlstr string, params interface{}, resObj interface{}, optionalHeader ...map[string]string) error {
	u, err := url.Parse(urlstr)
	if err != nil {
		return err
	}
	if params != nil {
		ps := httpclient.SimpleKVToQs(params)
		qs := u.Query()
		for k := range ps {
			qs.Add(k, ps.Get(k))
		}
		u.RawQuery = qs.Encode()
	}
	return jc.Get(u.String(), resObj, optionalHeader...)
}

// Get url
func (jc *JSONClient) Get(urlstr string, resObj interface{}, optionalHeader ...map[string]string) error {
	var header map[string]string
	if len(optionalHeader) > 0 {
		header = optionalHeader[0]
	}
	return jc.HttpRequest("GET", urlstr, header, nil, resObj)
}

// Post url
func (jc *JSONClient) Post(urlstr string, payload interface{}, resObj interface{}, optionalHeader ...map[string]string) error {
	var data []byte
	switch v := payload.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case nil:
		// do nothing
	default:
		var err error
		if data, err = sysjson.Marshal(payload); err != nil {
			return err
		}
	}
	header := map[string]string{"Content-Type": "application/json"}
	if len(optionalHeader) > 0 {
		for k, v := range optionalHeader[0] {
			header[k] = v
		}
	}
	return jc.HttpRequest("POST", urlstr, header, data, resObj)
}

// HttpRequest do http request
func (jc *JSONClient) HttpRequest(method, urlstr string, headers map[string]string, bodyData []byte, resObj interface{}) error {
	var abandonRes bool
	if resObj == nil {
		abandonRes = true
	} else {
		if reflect.ValueOf(resObj).Kind() != reflect.Ptr {
			return errors.New("res obj must be pointer")
		}
	}
	res, err := httpclient.HttpRequest(jc, method, urlstr, headers, bodyData)
	if err != nil {
		return err
	}
	if !abandonRes {
		if err = sysjson.Unmarshal(res, &resObj); err != nil {
			return err
		}
	}
	return nil
}
