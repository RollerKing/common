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
	"strings"
	"time"
)

type IHTTPDebugger interface {
	IsDebugOn() bool
	Print(tm time.Time, state string, title string, detail interface{})
}

type IHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	IHTTPDebugger
}

type Header map[string]string
type Form map[string]interface{}

func Get(c IHTTPClient, uri string) (res []byte, err error) {
	return HttpRequest(c, "GET", uri, nil, nil)
}

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

// type of data can be struct/map or json string/bytes
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

func HttpRequest(c IHTTPClient, method, urlstr string, headers Header, bodyData []byte) (res []byte, err error) {
	if c.IsDebugOn() {
		tm := time.Now()
		defer func() {
			c.Print(tm, "BEGIN", method, urlstr)
			c.Print(tm, "BEGIN", "HEADER", headers)
			c.Print(tm, "BEGIN", "BODY", string(bodyData))
			endat := time.Now()
			c.Print(endat, "END", "RESPONSE", string(res))
			c.Print(endat, "END", "COST", endat.Sub(tm))
		}()
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
	req, err = http.NewRequest(method, urlstr, body_data)
	if err != nil {
		return res, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rs, err := c.Do(req)
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

type Debugger struct {
	DebugOn bool
}

func (d *Debugger) IsDebugOn() bool {
	return d.DebugOn
}

func (d *Debugger) Print(tm time.Time, state string, title string, detail interface{}) {
	switch title {
	case "HEADER":
		var dh string
		if detail != nil {
			for k, v := range detail.(Header) {
				dh += k + ":" + v + "  "
			}
		}
		fmt.Printf("%v %s %s %s\n", tm.Format("2006-01-02 15:04:05"), state, title, dh)
	case "BODY":
		fmt.Printf("%v %s %s %s\n", tm.Format("2006-01-02 15:04:05"), state, title, detail)
	case "RESPONSE":
		fmt.Printf("%v %s %s %s\n", tm.Format("2006-01-02 15:04:05"), state, title, detail)
	case "COST":
		fmt.Printf("%v %s %s %s\n", tm.Format("2006-01-02 15:04:05"), state, title, detail.(time.Duration).String())
	default:
		fmt.Printf("%v %s %s %v\n", tm.Format("2006-01-02 15:04:05"), state, title, detail)
	}
}

type SimpleClient struct {
	*http.Client
	*Debugger
}

func GetSimpleClient(hc *http.Client) *SimpleClient {
	return &SimpleClient{
		Client:   hc,
		Debugger: &Debugger{},
	}
}

func (sc *SimpleClient) DoRequest(method, urlstr string, headers Header, bodyData []byte) (res []byte, err error) {
	return HttpRequest(sc, method, urlstr, headers, bodyData)
}
