package actions

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type EmailSender struct {
	User     string
	Password string
	Host     string
	Port     string
	Reply    string
}

func NewEmailSender() *EmailSender {
	return &EmailSender{
		User:     smtpConfig.User,
		Password: smtpConfig.Password,
		Host:     smtpConfig.Host,
		Port:     smtpConfig.Port,
		Reply:    smtpConfig.ReplyEmailAddress,
	}
}

func (s *EmailSender) Send(toAddress, subject, content string) {
	var err error
	if s.IsTls465() {
		err = s.handleTls465(toAddress, subject, content)
	} else {
		err = s.handleOthers(toAddress, subject, content)
	}
	if err != nil {
		logger.Errorf("EmailSender Send err : %s", err.Error())
	}
}

func (s *EmailSender) handleOthers(toAddress, subject, content string) error {
	user := s.User
	password := s.Password
	host := fmt.Sprintf("%s:%s", s.Host, s.Port)

	to := []string{toAddress}
	cc := []string{}
	bcc := []string{}

	mailType := "text"
	replyToAddress := s.Reply

	body := content

	if err := SendToMail(user, password, host, subject, body, mailType, replyToAddress, to, cc, bcc); err != nil {
		return errors.New("send email error: " + err.Error())
	} else {
		return nil
	}
}

func (s *EmailSender) handleTls465(toAddress, subject, content string) error {
	from := mail.Address{"", s.Reply}
	to := mail.Address{"", toAddress}
	subj := subject
	body := content

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := fmt.Sprintf("%s:%s", s.Host, s.Port)

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", s.User, s.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()

	return nil
}

func (s *EmailSender) IsTls465() bool {
	return s.Port == "465"
}

func SendToMail(user, password, host, subject, body, mailtype, replyToAddress string, to, cc, bcc []string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	cc_address := strings.Join(cc, ";")
	bcc_address := strings.Join(bcc, ";")
	to_address := strings.Join(to, ";")
	msg := []byte("To: " + to_address + "\r\nFrom: " + user + "\r\nSubject: " + subject + "\r\nReply-To: " + replyToAddress + "\r\nCc: " + cc_address + "\r\nBcc: " + bcc_address + "\r\n" + content_type + "\r\n\r\n" + body)

	send_to := MergeSlice(to, cc)
	send_to = MergeSlice(send_to, bcc)
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func MergeSlice(s1 []string, s2 []string) []string {
	slice := make([]string, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}
