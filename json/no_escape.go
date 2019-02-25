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
	data := buf.Bytes()
	// drop extra \n byte
	if size := len(data); size > 0 && data[size-1] == byte(10) {
		data = data[:size-1]
	}
	return data, nil
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

// UnsafeMarshalString marshal without error
func UnsafeMarshalString(v interface{}) string {
	data, err := Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

// UnsafeMarshalIndent marshal without error
func UnsafeMarshalIndent(v interface{}) []byte {
	data, err := Marshal(v)
	if err != nil {
		return []byte("")
	}
	var out bytes.Buffer
	sysjson.Indent(&out, data, "", "\t")
	return out.Bytes()
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
