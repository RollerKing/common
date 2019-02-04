package aux

import (
	"crypto/md5"
	"fmt"
)

// Md5 md5 string
func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
