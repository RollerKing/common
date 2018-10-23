package myip

import (
	"errors"
	"github.com/qjpcpu/common/web"
	"github.com/qjpcpu/common/web/json"
	"strings"
)

var ip_address string

func ResolvePublicIP() string {
	if ip_address == "" {
		ip_address = resolvePublicIP()
	}
	return ip_address
}

func RefreshIP() string {
	ip_address = ResolvePublicIP()
	return ip_address
}

func resolvePublicIP() string {
	if ip, err := getIPFromTaobao(); err == nil {
		return ip
	}
	if ip, err := getIPFromIPCN(); err == nil {
		return ip
	}
	return "127.0.0.1"
}

func getIPFromTaobao() (string, error) {
	var res struct {
		Data struct {
			Ip string `json:"ip"`
		} `json:"data"`
	}
	if err := json.Get("http://ip.taobao.com/service/getIpInfo2.php?ip=myip", &res); err != nil {
		return "", err
	}
	if res.Data.Ip == "" {
		return "", errors.New("can't get ip via taobao")
	}
	return res.Data.Ip, nil
}

func getIPFromIPCN() (string, error) {
	client := web.NewClient()
	client.SetHeaders(web.Header{"User-Agent": "curl/7.54.0"})
	data, err := client.Get("http://www.ip.cn")
	if err != nil {
		return "", err
	}
	str := strings.TrimPrefix(string(data), `当前 IP：`)
	return strings.Split(str, " ")[0], nil
}
