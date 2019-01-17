package email

import (
	"net/smtp"
	"strings"
)

// 腾讯企业邮
type ExmailQQ struct {
	server, port            string
	Account, Password, Name string
}

// 新建腾讯企业邮
func NewExmailQQ(account, password, name string) ExmailQQ {
	if name == "" {
		name = strings.Split(account, "@")[0]
	}
	return ExmailQQ{
		server:   "smtp.exmail.qq.com",
		port:     ":25",
		Account:  account,
		Password: password,
		Name:     name,
	}

}

// 发送邮件
func (eq ExmailQQ) SendMail(subject string, to []string, data []byte) error {
	auth := smtp.PlainAuth("", eq.Account, eq.Password, eq.server)
	contentType := "Content-Type: text/html; charset=UTF-8\r\n"
	msg := contentType + "From: " + eq.Name + "<" + eq.Account + ">\r\n" +
		"To: " + strings.Join(to, ",") + "\r\n" +
		"Subject: " + subject + "\r\n" + "\r\n"

	err := smtp.SendMail(eq.server+eq.port, auth, eq.Account, to, append([]byte(msg), data...))
	return err
}
