package smtp

import (
	"fmt"
	"net/smtp"
)

const smtpHost = "smtp.gmail.com"
const smtpPort = "587"

type Sender interface {
	SendEmail(to, subject, body string) error
}

type SenderImp struct {
	from string
	auth smtp.Auth
}

func NewSMTPHandler(from, password string) Sender {
	var ans SenderImp
	ans.from = from
	ans.auth = smtp.PlainAuth("", from, password, smtpHost)
	return ans
}
func (s SenderImp) SendEmail(to, subject, body string) error {
	receivers := []string{to}
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	msg := []byte(fmt.Sprintf("to: %s\nsubject: %s\n%s\n\n%s", to, subject, mime, body))

	return smtp.SendMail(smtpHost+":"+smtpPort, s.auth, s.from, receivers, msg)

}
