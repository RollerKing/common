package json

import (
	"bufio"
	"bytes"
	sysjson "encoding/json"
	"github.com/qjpcpu/go-prettyjson"
)

type RawMessage = sysjson.RawMessage

// PrettyMarshal colorful json
func PrettyMarshal(v interface{}) []byte {
	data, _ := prettyjson.Marshal(v)
	return data
}

// Marshal disable html escape
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	encoder := sysjson.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal same as sys unmarshal
func Unmarshal(data []byte, v interface{}) error {
	decoder := sysjson.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}

// MustMarshal must marshal successful
func MustMarshal(v interface{}) []byte {
	data, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// UnsafeMarshal marshal without error
func UnsafeMarshal(v interface{}) []byte {
	data, err := Marshal(v)
	if err != nil {
		return []byte("")
	}
	return data
}

// MustUnmarshal must unmarshal successful
func MustUnmarshal(data []byte, v interface{}) {
	if err := Unmarshal(data, v); err != nil {
		panic(err)
	}
}

// DecodeJSONP 剔除jsonp包裹层
func DecodeJSONP(str []byte) []byte {
	var start, end int
	for i := 0; i < len(str); i++ {
		if str[i] == '(' {
			start = i
			break
		}
	}
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == ')' {
			end = i
			break
		}
	}
	if end > 0 {
		return str[start+1 : end]
	} else {
		return str
	}
}
