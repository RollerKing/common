package httpclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/qjpcpu/common/crumbs"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

// HTTPError http error
type HTTPError struct {
	Code     int
	Status   string
	Response []byte
}

// Error http error content
func (he *HTTPError) Error() string {
	return fmt.Sprintf("%s\n%s", he.Status, he.Response)
}

// UnmarshalFunc data bytes to object pointer
type UnmarshalFunc func([]byte, interface{}) error

// IHTTPInspector debugger
type IHTTPInspector interface {
	IsDebugOn() bool
	Inspect(TraceData)
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

// GetWithParams get with qs object(map or struct)
func GetWithParams(c IHTTPClient, uri string, params interface{}, extraHeader Header) ([]byte, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if params != nil {
		ps := SimpleKVToQs(params)
		qs := u.Query()
		for k := range ps {
			qs.Add(k, ps.Get(k))
		}
		u.RawQuery = qs.Encode()
	}
	return Get(c, u.String(), extraHeader)
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

// Resolve response
func Resolve(data []byte, err error, resPtr interface{}, fn UnmarshalFunc) error {
	if err != nil {
		return err
	}
	return fn(data, resPtr)
}

// HttpRequest http req
func HttpRequest(c IHTTPClient, method, urlstr string, headers Header, bodyData []byte) (res []byte, err error) {
	var req *http.Request
	var bodyReader io.Reader
	var rs *http.Response
	if c.IsDebugOn() {
		tm := time.Now()
		defer func() {
			trdata := buildTraceData(urlstr, req, rs, bodyData, res, tm)
			c.Inspect(trdata)
		}()
	}
	method = strings.ToUpper(method)
	if !strings.HasPrefix(urlstr, "http://") && !strings.HasPrefix(urlstr, "https://") {
		urlstr = "http://" + urlstr
	}
	if len(bodyData) > 0 {
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
	if rs.StatusCode >= http.StatusBadRequest {
		return res, &HTTPError{
			Code:     rs.StatusCode,
			Status:   rs.Status,
			Response: res,
		}
	}
	return res, nil
}

// HttpStream send http stream
func HttpStream(c IHTTPClient, method, urlstr string, headers Header, bodyReader io.Reader) (res []byte, err error) {
	var req *http.Request
	var rs *http.Response
	if c.IsDebugOn() {
		tm := time.Now()
		defer func() {
			trdata := buildTraceData(urlstr, req, rs, nil, res, tm)
			c.Inspect(trdata)
		}()
	}
	if !strings.HasPrefix(urlstr, "http://") && !strings.HasPrefix(urlstr, "https://") {
		urlstr = "http://" + urlstr
	}
	method = strings.ToUpper(method)
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
	if rs.StatusCode >= http.StatusBadRequest {
		return res, &HTTPError{
			Code:     rs.StatusCode,
			Status:   rs.Status,
			Response: res,
		}
	}
	return res, nil
}

// Debugger debugger
type Debugger struct {
	DebugOn bool
	writer  io.Writer
}

// TraceData data
type TraceData struct {
	GoroutineID string
	Method      string
	URL         string
	ReqHeader   http.Header
	ReqPayload  []byte
	ReqAt       time.Time
	Cost        time.Duration
	StatusCode  int
	Status      string
	ResHeader   http.Header
	ResData     []byte
}

func buildTraceData(uri string, req *http.Request, res *http.Response, payload, body []byte, reqAt time.Time) TraceData {
	tr := TraceData{
		GoroutineID: crumbs.GetGoroutineID(),
		URL:         uri,
		ReqAt:       reqAt,
		Cost:        time.Since(reqAt),
		ReqPayload:  payload,
		ResData:     body,
	}
	if req != nil {
		if req.URL != nil {
			tr.URL = req.URL.String()
		}
		tr.Method = req.Method
		tr.ReqHeader = req.Header
	}
	if res != nil {
		tr.Status = res.Status
		tr.StatusCode = res.StatusCode
		tr.ResHeader = res.Header
	}
	return tr
}

// NewDebugger new simple debugger
func NewDebugger() *Debugger {
	return &Debugger{
		DebugOn: false,
		writer:  os.Stderr,
	}
}

// SetWriter set output
func (d *Debugger) SetWriter(w io.Writer) *Debugger {
	d.writer = w
	return d
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
func (d *Debugger) Inspect(tr TraceData) {
	writer := d.writer
	if writer == nil {
		return
	}
	var reqHeaders, resHeaders []string
	for k := range tr.ReqHeader {
		reqHeaders = append(reqHeaders, k+"="+tr.ReqHeader.Get(k))
	}
	for k := range tr.ResHeader {
		resHeaders = append(resHeaders, k+"="+tr.ResHeader.Get(k))
	}
	reqBody := string(tr.ReqPayload)
	resBody := string(tr.ResData)
	if reqBody != "" {
		reqBody = "\n" + reqBody
	}
	if resBody != "" {
		resBody = "\n" + resBody + "\n"
	}
	fmt.Fprintf(
		writer,
		"[%s] %s %s\n[statistics]: grtid=%s, reqat=%s, cost=%v \n[req headers]: %s\n[req body]:%s\n[res headers]: %s\n[response]:%s\n",
		tr.Method,
		tr.URL,
		tr.Status,
		tr.GoroutineID,
		tr.ReqAt.Format("2006-01-02 15:04:05.000"),
		tr.Cost,
		strings.Join(reqHeaders, " "),
		reqBody,
		strings.Join(resHeaders, " "),
		resBody,
	)
}

// SimpleKVToQs simple struct or map to url values
func SimpleKVToQs(obj interface{}, optTimeLayout ...string) url.Values {
	tp := reflect.TypeOf(obj)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	timeLayout := "2006-01-02 15:04:05"
	if len(optTimeLayout) > 0 && optTimeLayout[0] != "" {
		timeLayout = optTimeLayout[0]
	}
	switch tp.Kind() {
	case reflect.Map:
		return maptoQs(obj, timeLayout)
	case reflect.Struct:
		return structToQs(obj, timeLayout)
	}
	return url.Values{}
}

var typeOfTime = reflect.TypeOf(time.Time{})

func structToQs(obj interface{}, timeLayout string) url.Values {
	vals := url.Values{}
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	num := value.NumField()
	for i := 0; i < num; i++ {
		valueField := value.Field(i)
		typeField := value.Type().Field(i)
		if valueField.CanInterface() && valueField.Interface() != nil {
			var kstr, vstr string
			tp := typeField.Type
			if tp.Kind() == reflect.Ptr {
				tp = typeField.Type.Elem()
				valueField = valueField.Elem()
			}
			if tp.Kind() == reflect.Struct && tp == typeOfTime {
				if valueField.IsValid() {
					vstr = valueField.Interface().(time.Time).Format(timeLayout)
				}
			} else if tp.Kind() == reflect.Slice || tp.Kind() == reflect.Array {
				size := valueField.Len()
				list := make([]string, size)
				for i := 0; i < size; i++ {
					if tp.Elem() == typeOfTime {
						list[i] = valueField.Index(i).Interface().(time.Time).Format(timeLayout)
					} else {
						list[i] = fmt.Sprint(valueField.Index(i).Interface())
					}
				}
				vstr = strings.Join(list, ",")
			} else {
				vstr = fmt.Sprint(valueField.Interface())
			}
			var tag string
			for getTag := true; getTag; getTag = false {
				if tag = strings.Split(typeField.Tag.Get("form"), ",")[0]; tag != "" {
					break
				}
				if tag = strings.Split(typeField.Tag.Get("json"), ",")[0]; tag != "" {
					break
				}
			}
			if tag != "" {
				kstr = tag
			} else {
				kstr = crumbs.UnderlineLowercase(typeField.Name)
			}
			if vstr != "" {
				vals.Add(kstr, vstr)
			}
		}
	}
	return vals
}

func maptoQs(hash interface{}, timeLayout string) url.Values {
	vals := url.Values{}
	tp := reflect.TypeOf(hash)
	value := reflect.ValueOf(hash)
	if tp.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	keys := value.MapKeys()
	for _, key := range keys {
		vv := value.MapIndex(key)
		if !key.CanInterface() || !vv.CanInterface() || vv.Interface() == nil {
			continue
		}
		keyStr := fmt.Sprint(key.Interface())
		var vStr string
		vtp := reflect.TypeOf(vv.Interface())
		switch vtp.Kind() {
		case reflect.Struct:
			if vtp == typeOfTime {
				vStr = vv.Interface().(time.Time).Format(timeLayout)
			}
		case reflect.Slice, reflect.Array:
			vvv := reflect.ValueOf(vv.Interface())
			size := vvv.Len()
			list := make([]string, size)
			for i := 0; i < size; i++ {
				if vtp.Elem() == typeOfTime {
					list[i] = vvv.Index(i).Interface().(time.Time).Format(timeLayout)
				} else {
					list[i] = fmt.Sprint(vvv.Index(i).Interface())
				}
			}
			vStr = strings.Join(list, ",")
		default:
			vStr = fmt.Sprint(vv.Interface())
		}
		if vStr != "" {
			vals.Add(keyStr, vStr)
		}
	}
	return vals
}
