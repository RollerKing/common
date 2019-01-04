package jsonhttp

import (
	"bufio"
	"bytes"
	sysjson "encoding/json"
	"errors"
	"github.com/qjpcpu/go-prettyjson"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"strings"
	"time"
)

var client = NewClient(false)

type JSONClient struct {
	*http.Client
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

func Colored(obj interface{}) string {
	s, _ := prettyjson.Marshal(obj)
	return string(s)
}

func Get(urlstr string, resObj interface{}, optional_header ...map[string]string) error {
	return client.Get(urlstr, resObj, optional_header...)
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
	var req *http.Request
	var body_data io.Reader
	method = strings.ToUpper(method)
	if !strings.HasPrefix(urlstr, "http://") && !strings.HasPrefix(urlstr, "https://") {
		urlstr = "http://" + urlstr
	}
	if bodyData != nil && len(bodyData) > 0 {
		body_data = bytes.NewBuffer(bodyData)
	}
	req, err := http.NewRequest(method, urlstr, body_data)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rs, err := jc.Do(req)
	if err != nil {
		return err
	}
	defer rs.Body.Close()
	res, err := ioutil.ReadAll(rs.Body)
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
