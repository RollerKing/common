package crumbs

import (
	"crypto/md5"
	"fmt"
)

// MD5 md5 string
func MD5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
