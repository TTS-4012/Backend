package smtp

import (
	"fmt"
	"net/smtp"
	"ocontest/pkg"
	"ocontest/pkg/configs"
)

const smtpHost = "smtp.gmail.com"
const smtpPort = "587"

type Sender interface {
	SendEmail(to, subject, body string) error
}

type SenderImp struct {
	from    string
	auth    smtp.Auth
	enabled bool
}

func NewSMTPHandler(c configs.SectionSMTP) Sender {
	var ans SenderImp
	if c.Enabled {
		ans.from = c.From
		ans.auth = smtp.PlainAuth("", c.From, c.Password, smtpHost)
		ans.enabled = true
	} else {
		pkg.Log.Warning("smtp is not enabled, this should only be in dev environments, in production it must be enabled")
		ans.enabled = false // not necessary, only for precaution
	}
	return ans
}
func (s SenderImp) SendEmail(to, subject, body string) error {
	receivers := []string{to}
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	msg := []byte(fmt.Sprintf("to: %s\nsubject: %s\n%s\n\n%s", to, subject, mime, body))
	if !s.enabled {
		pkg.Log.Info("SendEmail in smtp disable mode, logging message: \n", string(msg))
		return nil
	}

	return smtp.SendMail(smtpHost+":"+smtpPort, s.auth, s.from, receivers, msg)

}
