/*
微信企业支付
*/
package pay

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/qjpcpu/common/myip"
	"github.com/qjpcpu/log"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// 支付文档地址: https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_2
type CheckNameOption string

const (
	NO_CHECK    CheckNameOption = "NO_CHECK"
	FORCE_CHECK                 = "FORCE_CHECK"
)

type WeixinPayArgs struct {
	XMLName xml.Name `xml:"xml" json:"-"`
	// 商户账号appid
	AppID string `xml:"mch_appid" json:"mch_appid"`
	// 商户号
	MchID string `xml:"mchid" json:"mchid"`
	// 随机字符串
	Nonce string `xml:"nonce_str" json:"nonce_str"`
	// 签名
	Sign string `xml:"sign" json:"sign"`
	// 商户订单号
	TradeID string `xml:"partner_trade_no" json:"partner_trade_no"`
	// 用户openid
	OpenID string `xml:"openid" json:"openid"`
	// 校验用户姓名选项
	CheckName CheckNameOption `xml:"check_name" json:"check_name"`
	// 金额,企业付款金额，单位为分
	Amount int64 `xml:"amount" json:"amount"`
	// 企业付款描述信息
	Desc string `xml:"desc" json:"desc"`
	// 该IP同在商户平台设置的IP白名单中的IP没有关联，该IP可传用户端或者服务端的IP
	IP string `xml:"spbill_create_ip" json:"spbill_create_ip"`

	// optional:
	// 收款用户姓名
	UserName string `xml:"re_user_name,omitempty" json:"re_user_name,omitempty"`
}

type WeixinPayClient struct {
	transport *http.Transport
	appid     string
	appsecret string
	mchid     string
	secretkey string
}

func (client *WeixinPayClient) CreatePayArgs(tradeId, openid, desc string, amount int64) *WeixinPayArgs {
	return &WeixinPayArgs{
		AppID:     client.appid,
		MchID:     client.mchid,
		Nonce:     GenSimpleUniqueId(),
		TradeID:   tradeId,
		OpenID:    openid,
		CheckName: NO_CHECK,
		Amount:    amount,
		Desc:      desc,
		IP:        myip.ResolvePublicIP(),
	}
}

type WeixinPayRes struct {
	XMLName          xml.Name `xml:"xml" json:"-"`
	Return_code      string   `xml:"return_code"`
	Return_msg       string   `xml:"return_msg"`
	Mch_appid        string   `xml:"mch_appid"`
	Mchid            string   `xml:"mchid"`
	Result_code      string   `xml:"result_code"`
	Err_code         string   `xml:"err_code"`
	Err_code_des     string   `xml:"err_code_des"`
	Partner_trade_no string   `xml:"partner_trade_no"`
	Payment_no       string   `xml:"payment_no"`
	Payment_time     string   `xml:"payment_time"`
}

type KeyPairLoader func() (tls.Certificate, error)

func KeyPairContentLoader(certBytes, keyBytes []byte) KeyPairLoader {
	return func() (tls.Certificate, error) {
		return tls.X509KeyPair(certBytes, keyBytes)
	}
}

func KeyPairFileLoader(certfile, keyfile string) KeyPairLoader {
	return func() (tls.Certificate, error) {
		return tls.LoadX509KeyPair(certfile, keyfile)
	}
}

// 企业支付
func InitWeixinPayClient(secretkey, appid, appsecret, mchid string, keypairLoader KeyPairLoader) (*WeixinPayClient, error) {
	if secretkey == "" || appid == "" || appsecret == "" || mchid == "" {
		return nil, errors.New("parameters err")
	}
	cliCrt, err := keypairLoader()
	if err != nil {
		return nil, err
	}
	client := &WeixinPayClient{
		transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cliCrt},
			},
		},
		appid:     appid,
		appsecret: appsecret,
		secretkey: secretkey,
		mchid:     mchid,
	}
	return client, nil
}

func (args WeixinPayArgs) XML() []byte {
	data, _ := xml.Marshal(args)
	return data
}

func (args WeixinPayArgs) Validate() error {
	var err error
	for loop := true; loop; loop = false {
		if args.AppID == "" {
			err = errors.New("no appid")
			break
		}
		if args.MchID == "" {
			err = errors.New("no mchid")
			break
		}
		if args.Nonce == "" {
			err = errors.New("no nonce_str")
			break
		}
		if args.TradeID == "" {
			err = errors.New("no trade id")
			break
		}
		if args.OpenID == "" {
			err = errors.New("no openid")
			break
		}
		if args.CheckName != NO_CHECK && args.CheckName != FORCE_CHECK {
			err = errors.New("bad check name option")
			break
		}
		if args.CheckName == FORCE_CHECK && args.UserName == "" {
			err = errors.New("check_name设置为FORCE_CHECK，则必填用户真实姓名")
			break
		}
		if args.Amount < 100 {
			err = fmt.Errorf("金额错误:目前最低付款金额为1元，最高200w,当前:%v元", args.Amount/100)
			break
		}
		if args.Desc == "" {
			err = errors.New("no description")
			break
		}
		if args.IP == "" {
			err = errors.New("no client ip")
			break
		}
	}
	return err
}

func (args *WeixinPayArgs) CalcSign(mchkey string) (string, error) {
	// 第一步，设所有发送或者接收到的数据为集合M，将集合M内非空参数值的参数按照参数名ASCII码从小到大排序（字典序），使用URL键值对的格式（即key1=value1&key2=value2…）拼接成字符串stringA
	data, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	kv := make(map[string]interface{})
	if err = json.Unmarshal(data, &kv); err != nil {
		return "", err
	}
	var keys []string
	for k, v := range kv {
		if k == "sign" {
			continue
		}
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		keys = append(keys, k)
	}
	tokens := make([]string, len(keys))
	sort.Strings(keys)
	for i, k := range keys {
		tokens[i] = fmt.Sprintf(`%s=%v`, k, kv[k])
	}
	stringA := strings.Join(tokens, "&")
	// 在stringA最后拼接上key得到stringSignTemp字符串，并对stringSignTemp进行MD5运算，再将得到的字符串所有字符转换为大写，得到sign值signValue
	// key为商户平台设置的密钥key,key设置路径：微信商户平台(pay.weixin.qq.com)-->账户设置-->API安全-->密钥设置
	stringSignTemp := stringA + "&key=" + mchkey
	sign := fmt.Sprintf("%X", md5.Sum([]byte(stringSignTemp)))
	args.Sign = sign
	// mac := hmac.New(sha256.New, []byte(mchkey))
	// mac.Write([]byte(sign))
	// sign = fmt.Sprintf("%X", mac.Sum(nil))
	return sign, nil
}

func (wxclient *WeixinPayClient) Pay(args *WeixinPayArgs) (string, error) {
	if err := args.Validate(); err != nil {
		return "", err
	}
	if _, err := args.CalcSign(wxclient.secretkey); err != nil {
		return "", err
	}
	if wxclient.transport == nil {
		return "", errors.New("transport not initialized")
	}
	data := args.XML()
	log.Debugf("发送的请求payload:\n%s", string(data))
	buf := bytes.NewBuffer(data)
	client := &http.Client{
		Transport: wxclient.transport,
	}
	resp, err := client.Post(`https://api.mch.weixin.qq.com/mmpaymkttransfers/promotion/transfers`, "application/xml", buf)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s", resp.Status)
		return "", err
	}
	defer resp.Body.Close()

	rtData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Debugf("返回:\n%s", string(rtData))

	takeout_res := WeixinPayRes{}
	err = xml.Unmarshal([]byte(rtData), &takeout_res)
	if err != nil {
		return "", err
	}
	if takeout_res.Return_code != "SUCCESS" ||
		takeout_res.Result_code != "SUCCESS" {
		err = fmt.Errorf("%s: %s", takeout_res.Return_msg, takeout_res.Err_code_des)
		return takeout_res.Err_code, err
	}
	return "", nil
}
