package debug

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func WriteFile(filename string, data []byte) {
	dir := filepath.Dir(filename)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	} else {
		ShouldBeNil(err)
	}
	ShouldBeNil(ioutil.WriteFile(filename, data, 0644))
}

func ReadFile(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	ShouldBeNil(err)
	return data
}
