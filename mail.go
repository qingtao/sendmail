//Package sendmail use net/smtp to send email
package sendmail

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/smtp"
)

var (
	//ErrNoMailFrom no from
	ErrNoMailFrom = errors.New("from field is empty")
	//ErrNoMailTo no to
	ErrNoMailTo = errors.New("to field is empty")
)

//Join join slice of string with separator ","
func Join(a []string, sep string) string {
	s := ""
	for i := 0; i < len(a); i++ {
		if i == 0 {
			s = a[i]
			continue
		}
		s += "," + a[i]
	}
	return s
}

//Sendmail use smtp.SendMail
func Sendmail(addr, from, password string, to, cc, bcc []string, sub, msg []byte) error {
	//check addr with net.SplitHostPort
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	if from == "" {
		return ErrNoMailFrom
	}

	if len(to) < 1 {
		return ErrNoMailTo
	}

	var b bytes.Buffer
	//add Subject
	b.WriteString(fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", base64.StdEncoding.EncodeToString(sub)))
	//add From
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	//add To
	toHeader := Join(to, ",")
	//rcpt to
	b.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))

	//Cc
	ccHeader := Join(cc, ",")
	if ccHeader != "" {
		b.WriteString(fmt.Sprintf("Cc: %s\r\n", ccHeader))
		to = append(to, cc...)
	}

	//Bcc
	bccHeader := Join(bcc, ",")
	if bccHeader != "" {
		b.WriteString(fmt.Sprintf("Bcc: %s\r\n", bccHeader))
		to = append(to, bcc...)
	}

	//add charset: UTF-8
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	//add encoding: base64
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	//add \r\n
	b.WriteString("\r\n")
	b.WriteString(base64.StdEncoding.EncodeToString(msg))
	//add \r\n
	b.WriteString("\r\n")
	if password == "" {
		if err := smtp.SendMail(addr, nil, from, to, b.Bytes()); err != nil {
			return err
		}
	} else {
		//PlainAuth
		a := smtp.PlainAuth("", from, password, host)
		if err := smtp.SendMail(addr, a, from, to, b.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

//SendmailSkipTLS sendmail without TLS verify, it only should be used on localhost
func SendmailSkipTLS(addr, from, password string, to, cc, bcc []string, sub, msg []byte) error {
	//check addr with net.SplitHostPort
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	if from == "" {
		return ErrNoMailFrom
	}

	if len(to) < 1 {
		return ErrNoMailTo
	}

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	//skip tls verify
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	if err = c.StartTLS(config); err != nil {
		return err
	}
	//auth password
	if password != "" {
		a := smtp.PlainAuth("", from, password, host)
		if err := c.Auth(a); err != nil {
			return err
		}
	}

	var b bytes.Buffer
	//add Subject
	b.WriteString(fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", base64.StdEncoding.EncodeToString(sub)))

	if err = c.Mail(from); err != nil {
		return err
	}
	//add From
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))

	//add To
	toHeader := Join(to, ",")
	//rcpt to
	b.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))

	//Cc
	ccHeader := Join(cc, ",")
	if ccHeader != "" {
		b.WriteString(fmt.Sprintf("Cc: %s\r\n", ccHeader))
		to = append(to, cc...)
	}

	//Bcc
	bccHeader := Join(bcc, ",")
	if bccHeader != "" {
		b.WriteString(fmt.Sprintf("Bcc: %s\r\n", bccHeader))
		to = append(to, bcc...)
	}

	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}

	//add charset: UTF-8
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	//add encoding: base64
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	//add \r\n
	b.WriteString("\r\n")
	b.WriteString(base64.StdEncoding.EncodeToString(msg))
	//add \r\n
	b.WriteString("\r\n")

	//write email data
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	if _, err := b.WriteTo(wc); err != nil {
		return err
	}
	return c.Quit()
}
