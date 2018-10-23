package myip

import (
	"testing"
)

func TestGetIP(t *testing.T) {
	var ip string
	var err error
	if ip, err = getIPFromTaobao(); err != nil {
		t.Fatal(err)
	}
	t.Logf("get ip from taobao: [%s]", ip)
	if ip, err = getIPFromIPCN(); err != nil {
		t.Fatal(err)
	}
	t.Logf("get ip from ip.cn: [%s]", ip)
}
