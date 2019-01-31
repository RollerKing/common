package httpclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// IHTTPInspector debugger
type IHTTPInspector interface {
	IsDebugOn() bool
	SetDebug(bool)
	Inspect(uri string, req *http.Request, res *http.Response, body []byte, cost time.Duration)
}

// IHTTPRequester internal http executor
type IHTTPRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

// IHTTPClient do http request
type IHTTPClient interface {
	IHTTPRequester
	IHTTPInspector
}

// Header http header
type Header map[string]string

// Form http form
type Form map[string]interface{}

// Get get req
func Get(c IHTTPClient, uri string, extraHeader Header) (res []byte, err error) {
	return HttpRequest(c, "GET", uri, extraHeader, nil)
}

// PostForm post form
func PostForm(c IHTTPClient, urlstr string, data Form, extraHeader Header) (res []byte, err error) {
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
	return HttpRequest(c, "POST", urlstr, hder, []byte(values.Encode()))
}

// PostJSON type of data can be struct/map or json string/bytes
func PostJSON(c IHTTPClient, urlstr string, data interface{}, extraHeader Header) (res []byte, err error) {
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
		var buf bytes.Buffer
		writer := bufio.NewWriter(&buf)
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(data); err != nil {
			return nil, err
		}
		writer.Flush()
		payload = buf.Bytes()
	}
	return HttpRequest(c, "POST", urlstr, hder, payload)
}

// HttpRequest http req
func HttpRequest(c IHTTPClient, method, urlstr string, headers Header, bodyData []byte) (res []byte, err error) {
	var req *http.Request
	var bodyReader io.Reader
	var rs *http.Response
	if c.IsDebugOn() {
		tm := time.Now()
		defer func() {
			c.Inspect(urlstr, req, rs, res, time.Since(tm))
		}()
	}
	method = strings.ToUpper(method)
	if !strings.HasPrefix(urlstr, "http://") && !strings.HasPrefix(urlstr, "https://") {
		urlstr = "http://" + urlstr
	}
	if bodyData != nil && len(bodyData) > 0 {
		bodyReader = bytes.NewBuffer(bodyData)
	}
	req, err = http.NewRequest(method, urlstr, bodyReader)
	if err != nil {
		return res, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rs, err = c.Do(req)
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

// Debugger debugger
type Debugger struct {
	DebugOn bool
}

// IsDebugOn is debug on
func (d *Debugger) IsDebugOn() bool {
	return d.DebugOn
}

// SetDebug set debug on/off
func (d *Debugger) SetDebug(set bool) {
	d.DebugOn = set
}

// Inspect inspect http entity
func (d *Debugger) Inspect(uri string, req *http.Request, res *http.Response, body []byte, cost time.Duration) {
	var reqHeaders, resHeaders []string
	if req != nil {
		for k := range req.Header {
			reqHeaders = append(reqHeaders, k+"="+req.Header.Get(k))
		}
	}
	if res != nil {
		for k := range res.Header {
			resHeaders = append(resHeaders, k+"="+res.Header.Get(k))
		}
	}
	var status string
	if res != nil {
		status = res.Status
	}
	fmt.Printf("[%s] %s %s\n[cost]: %v\n[req headers]: %s\n[res headers]: %s\n[response]:\n%s\n", req.Method, uri, status, cost, strings.Join(reqHeaders, "; "), strings.Join(resHeaders, "; "), string(body))
}

func SimpleKVToQs(obj interface{}) url.Values {
	if mp, ok := obj.(map[string]interface{}); ok {
		return mapToQs(mp)
	}
	if mp, ok := obj.(*map[string]interface{}); ok {
		return mapToQs(*mp)
	}
	if mp, ok := obj.(map[string]string); ok {
		vals := url.Values{}
		for k, v := range mp {
			vals.Add(k, v)
		}
		return vals
	}
	return structToQs(obj)
}

func structToQs(obj interface{}) url.Values {
	vals := url.Values{}
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	num := value.NumField()
	for i := 0; i < num; i++ {
		valueField := value.Field(i)
		typeField := value.Type().Field(i)
		if valueField.Interface() != nil {
			var kstr, vstr string
			tp := typeField.Type
			if tp.Kind() == reflect.Ptr {
				tp = typeField.Type.Elem()
				valueField = valueField.Elem()
			}
			if tp.Kind() == reflect.Struct && tp.String() == "time.Time" {
				if valueField.IsValid() {
					vstr = valueField.Interface().(time.Time).Format("2006-01-02 15:04:05")
				}
			} else if tp.Kind() == reflect.Slice || tp.Kind() == reflect.Array {
				size := valueField.Len()
				list := make([]string, size)
				for i := 0; i < size; i++ {
					list[i] = fmt.Sprint(valueField.Index(i).Interface())
				}
				vstr = strings.Join(list, ",")
			} else {
				vstr = fmt.Sprint(valueField.Interface())
			}
			tag := strings.Split(typeField.Tag.Get("qs"), ",")[0]
			if tag != "" {
				kstr = tag
			} else {
				kstr = lowercase_underline(typeField.Name)
			}
			if vstr != "" {
				vals.Add(kstr, vstr)
			}
		}
	}
	return vals
}

func mapToQs(hash map[string]interface{}) url.Values {
	vals := url.Values{}
	for k, v := range hash {
		if v != nil {
			var vstr string
			tp := reflect.TypeOf(v)
			value := reflect.ValueOf(v)
			if tp.Kind() == reflect.Ptr {
				tp = tp.Elem()
				value = value.Elem()
			}
			if tp.Kind() == reflect.Struct && tp.String() == "time.Time" {
				vstr = value.Interface().(time.Time).Format("2006-01-02 15:04:05")
			} else if tp.Kind() == reflect.Slice || tp.Kind() == reflect.Array {
				size := value.Len()
				list := make([]string, size)
				for i := 0; i < size; i++ {
					list[i] = fmt.Sprint(value.Index(i).Interface())
				}
				vstr = strings.Join(list, ",")
			} else {
				vstr = fmt.Sprint(v)
			}
			if vstr != "" {
				vals.Add(k, vstr)
			}
		}
	}
	return vals
}

func lowercase_underline(name string) string {
	data := []byte(name)
	var res []byte
	var i int
	for i < len(data) {
		if data[i] >= 65 && data[i] <= 90 {
			start := i
			i++
			for i < len(data) {
				if data[i] < 65 || data[i] > 90 {
					break
				}
				i++
			}
			res = append(res, byte(95))
			if i < len(data) && i-start >= 2 {
				res = append(res, data[start:i-1]...)
				res = append(res, byte(95), data[i-1])
			} else {
				res = append(res, data[start:i]...)
			}
			continue
		}
		res = append(res, data[i])
		i++
	}
	if len(res) > 0 && res[0] == byte(95) {
		res = res[1:]
	}
	return strings.ToLower(string(res))
}
