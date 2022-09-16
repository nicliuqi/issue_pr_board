package utils

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/jordan-wright/email"
	"net/smtp"
	"os"
)

func SendVerifyEmail(receiver string, code string) error {
	username := os.Getenv("SMTP_USERNAME")
	passwd := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	em := email.NewEmail()
	em.From = username
	em.To = []string{receiver}
	em.Subject = "openEuler QuickIssue"
	htmlBody := fmt.Sprintf("<p>Dear user,</p></br>"+
		"<p>We have received your request to submit a quick issue to openEuler projects. Please ignore this email if it is not operated by yourself.</p>"+
		"<p>The following is the verification code</p>"+
		"<p><b>%s</b></p>"+
		"<p>The verification code is valid for 10 minutes. If it expires, you need to obtain it again.</p></br>"+
		"Have any questions or need help? Just create an issue to <a href='https://gitee.com/openeuler/infrastructure/issues'>Infrastructure</a> or send an email to infra@openeuler.org.", code)
	em.HTML = []byte(htmlBody)
	auth := smtp.PlainAuth("", username, passwd, host)
	err := em.Send(host+":"+port, auth)
	if err != nil {
		return err
	} else {
		logs.Info("Send verification code to", receiver)
		return err
	}
}
