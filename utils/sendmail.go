package utils

import (
	"bytes"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"html/template"
	"issue_pr_board/config"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

const (
	CommentAttentionTemplate     = "templates/email/comment_attention.tmpl"
	NewIssueNotifyTemplate       = "templates/email/new_issue_notify.tmpl"
	StateChangeAttentionTemplate = "templates/email/state_change_attention.tmpl"
	VerifyTemplate               = "templates/email/verify.tmpl"
)

type EmailParams struct {
	Body      string
	Code      string
	Commenter string
	Link      string
	Number    string
	Receiver  string
	Repo      string
	State     string
	Title     string
}

func SendVerifyEmail(ep EmailParams) error {
	subject := "openEuler QuickIssue"
	htmlBody := loadTemplate(VerifyTemplate, ep)
	err := sendEmail(ep.Receiver, subject, htmlBody)
	if err != nil {
		return err
	} else {
		logs.Info("Send verification code to", ep.Receiver)
		return err
	}
}

func SendCommentAttentionEmail(ep EmailParams) error {
	subject := fmt.Sprintf("openEuler QuickIssue: #%v %v", ep.Number, ep.Title)
	htmlBody := loadTemplate(CommentAttentionTemplate, ep)
	err := sendEmail(ep.Receiver, subject, htmlBody)
	if err != nil {
		logs.Error(err)
		return err
	} else {
		logs.Info("Send issue comment attention to", ep.Receiver)
		return nil
	}
}

func SendStateChangeAttentionEmail(ep EmailParams) error {
	subject := fmt.Sprintf("openEuler QuickIssue: #%v %v", ep.Number, ep.Title)
	htmlBody := loadTemplate(StateChangeAttentionTemplate, ep)
	err := sendEmail(ep.Receiver, subject, htmlBody)
	if err != nil {
		logs.Error(err)
		return err
	} else {
		logs.Info("Send issue state change attention to", ep.Receiver)
		return nil
	}
}

func SendNewIssueNotifyEmail(ep EmailParams) error {
	subject := fmt.Sprintf("Notice a new issue -#%v", ep.Number)
	htmlBody := loadTemplate(NewIssueNotifyTemplate, ep)
	err := sendEmail(ep.Receiver, subject, htmlBody)
	if err != nil {
		logs.Error(err)
		return err
	} else {
		logs.Info("Send new issue notification to", ep.Receiver)
		return nil
	}
}

func loadTemplate(path string, data interface{}) string {
	content, err := os.ReadFile(path)
	name := filepath.Base(path)
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		logs.Error(err)
		return ""
	}
	renderString, err := renderTemplate(tmpl, data)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return renderString
}

func renderTemplate(tmpl *template.Template, data interface{}) (string, error) {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template")
	}
	return buf.String(), nil
}

func sendEmail(receiver, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", config.AppConfig.SMTPUsername, config.AppConfig.SMTPPassword,
		config.AppConfig.SMTPHost)
	contentType := "Content-Type: text/html; charset=UTF-8"
	msg := []byte("To: " + receiver + "\r\nFrom: " + config.AppConfig.SMTPSender + ">\r\nSubject: " + subject + "\r\n" +
		contentType + "\r\n\r\n" + htmlBody)
	err := smtp.SendMail(fmt.Sprintf("%v:%v", config.AppConfig.SMTPHost, config.AppConfig.SMTPPort), auth,
		config.AppConfig.SMTPUsername, strings.Split(receiver, ";"), msg)
	if err != nil {
		return err
	}
	return nil
}
