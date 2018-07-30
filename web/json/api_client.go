package json

import (
	"bufio"
	"bytes"
	sysjson "encoding/json"
	"errors"
	"github.com/hokaccha/go-prettyjson"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: 60 * time.Second,
}

func Colored(obj interface{}) string {
	s, _ := prettyjson.Marshal(obj)
	return string(s)
}

func Get(urlstr string, resObj interface{}) error {
	return HttpRequest("GET", urlstr, nil, nil, resObj)
}

func Post(urlstr string, payload interface{}, resObj interface{}) error {
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
	return HttpRequest("POST", urlstr, header, data, resObj)
}

func HttpRequest(method, urlstr string, headers map[string]string, bodyData []byte, resObj interface{}) error {
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
	rs, err := client.Do(req)
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
