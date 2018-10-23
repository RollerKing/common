package pay

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func GenSimpleUniqueId() string {
	prev, _ := time.Parse("20060102150405", "20150701000000")

	byts := make([]byte, 4)
	rand.Read(byts)
	var x uint32
	binary.Read(bytes.NewBuffer(byts), binary.BigEndian, &x)

	str := fmt.Sprintf("%16d%-16d", x, (time.Now().UnixNano()-prev.UnixNano())/1000)
	return strings.Trim(str, " ")
}
