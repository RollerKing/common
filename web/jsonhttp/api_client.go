package jsonhttp

import (
	"bufio"
	"bytes"
	sysjson "encoding/json"
	"errors"
	"fmt"
	"github.com/qjpcpu/common/web/httpclient"
	"github.com/qjpcpu/go-prettyjson"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"time"
)

var client = NewClient(false)

type JSONClient struct {
	*http.Client
	httpclient.Debugger
}

func NewClient(cookie bool) *JSONClient {
	jc := &JSONClient{
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
	if cookie {
		jar, _ := cookiejar.New(nil)
		jc.Client.Jar = jar
	}
	return jc
}

func (jc *JSONClient) SetTimeout(tm time.Duration) {
	jc.Client.Timeout = tm
}

func (js *JSONClient) Do(req *http.Request) (*http.Response, error) {
	return js.Client.Do(req)
}

func Colored(obj interface{}) string {
	s, _ := prettyjson.Marshal(obj)
	return string(s)
}

func Get(urlstr string, resObj interface{}, optional_header ...map[string]string) error {
	return client.Get(urlstr, resObj, optional_header...)
}

func (jc *JSONClient) GetWithParams(urlstr string, params map[string]interface{}, resObj interface{}, optional_header ...map[string]string) error {
	u, _ := url.Parse(urlstr)
	qs := u.Query()
	for k, v := range params {
		qs.Add(k, fmt.Sprint(v))
	}
	u.RawQuery = qs.Encode()
	return jc.Get(u.String(), resObj, optional_header...)
}

func (jc *JSONClient) Get(urlstr string, resObj interface{}, optional_header ...map[string]string) error {
	var header map[string]string
	if len(optional_header) > 0 {
		header = optional_header[0]
	}
	return jc.HttpRequest("GET", urlstr, header, nil, resObj)
}

func Post(urlstr string, payload interface{}, resObj interface{}, optional_header ...map[string]string) error {
	return client.Post(urlstr, payload, resObj, optional_header...)
}

func (jc *JSONClient) Post(urlstr string, payload interface{}, resObj interface{}, optional_header ...map[string]string) error {
	var data []byte
	switch v := payload.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case nil:
		// do nothing
	default:
		var buf bytes.Buffer
		writer := bufio.NewWriter(&buf)
		encoder := sysjson.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(payload); err != nil {
			return err
		}
		writer.Flush()
		data = buf.Bytes()
	}
	header := map[string]string{"Content-Type": "application/json"}
	if len(optional_header) > 0 {
		for k, v := range optional_header[0] {
			header[k] = v
		}
	}
	return jc.HttpRequest("POST", urlstr, header, data, resObj)
}

func HttpRequest(method, urlstr string, headers map[string]string, bodyData []byte, resObj interface{}) error {
	return client.HttpRequest(method, urlstr, headers, bodyData, resObj)
}

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
		decoder := sysjson.NewDecoder(bytes.NewReader(res))
		decoder.UseNumber()
		if err = decoder.Decode(&resObj); err != nil {
			return err
		}
	}
	return nil
}
