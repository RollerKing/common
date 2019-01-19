package json

import (
	"bufio"
	"bytes"
	sysjson "encoding/json"
	"github.com/qjpcpu/go-prettyjson"
)

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

// MustUnmarshal must unmarshal successful
func MustUnmarshal(data []byte, v interface{}) {
	if err := Unmarshal(data, v); err != nil {
		panic(err)
	}
}
