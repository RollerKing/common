package pay

import (
	"testing"
)

func TestCash(t *testing.T) {
	client, err := InitWeixinPayClient("", "", "", "", KeyPairFileLoader("x.cert", "x.key"))
	if err != nil {
		t.Fatal(err)
	}
	args := client.CreatePayArgs("my0ordreid", "on1MQ5Xb1pTAlNWqMiWEvmn8pUcM", "desc", 100)
	if _, err := client.Pay(args); err != nil {
		t.Fatal(err)
	}
}
