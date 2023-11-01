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
	msg := []byte(fmt.Sprintf("to: %s\nsubject: %s\n\n%s", to, subject, body))

	return smtp.SendMail(smtpHost+":"+smtpPort, s.auth, s.from, receivers, msg)

}
