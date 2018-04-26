//Package sendmail use net/smtp to send email
package sendmail

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

var (
	//ErrNoMailFrom address of from user is empty
	ErrNoMailFrom = errors.New("from field is empty")
	//ErrNoMailTo address of to is empty
	ErrNoMailTo = errors.New("to field is empty")
)

func parseAddress(a []string) ([]*mail.Address, error) {
	addrs := make([]*mail.Address, len(a))
	for i := 0; i < len(a); i++ {
		addr, err := mail.ParseAddress(a[i])
		if err != nil {
			return nil, err
		}
		addrs[i] = addr
	}
	return addrs, nil
}

func newMessage(from *mail.Address, to, cc, bcc []*mail.Address, sub string, msg []byte) (string, error) {
	var buf strings.Builder
	// write subject
	buf.WriteString("Subject: ")
	buf.WriteString(mime.BEncoding.Encode("utf-8", sub))
	buf.WriteString("\r\n")
	//write mail from
	buf.WriteString("From: ")
	buf.WriteString(from.String())
	buf.WriteString("\r\n")
	// write rcpt to
	buf.WriteString("To: ")
	for i := 0; i < len(to); i++ {
		buf.WriteString(to[i].String())
		if i != len(to)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteString("\r\n")
	// write cc
	buf.WriteString("Cc: ")
	for i := 0; i < len(cc); i++ {
		buf.WriteString(cc[i].String())
		if i != len(to)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteString("\r\n")
	// write Bcc
	buf.WriteString("Bcc: ")
	for i := 0; i < len(bcc); i++ {
		buf.WriteString(bcc[i].String())
		if i != len(to)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteString("\r\n")
	// write content-type
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString(msg))
	buf.WriteString("\r\n")
	return buf.String(), nil
}

//Sendmail use smtp.SendMail
func Sendmail(addr, from, password string, to, cc, bcc []string, sub string, msg []byte) error {
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
	mailfrom, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}
	rcptto, err := parseAddress(to)
	if err != nil {
		return err
	}
	ccto, err := parseAddress(cc)
	if err != nil {
		return err
	}
	bccto, err := parseAddress(bcc)
	if err != nil {
		return err
	}

	content, err := newMessage(mailfrom, rcptto, ccto, bccto, sub, msg)
	if err != nil {
		return err
	}
	rcptto = append(rcptto, ccto...)
	rcptto = append(rcptto, bccto...)
	mailto := make([]string, len(rcptto))
	for i := 0; i < len(rcptto); i++ {
		mailto[i] = rcptto[i].Address
	}
	fmt.Println(mailfrom.Address)
	fmt.Println(password)

	if password == "" {
		if err := smtp.SendMail(addr, nil, mailfrom.Address, mailto, []byte(content)); err != nil {
			return err
		}
	} else {
		//PlainAuth
		a := smtp.PlainAuth("", mailfrom.Address, password, host)
		if err := smtp.SendMail(addr, a, mailfrom.Address, mailto, []byte(content)); err != nil {
			return err
		}
	}
	return nil
}

//SkipVerifyTLS sendmail skip TLS verify, it only should be used on localhost
func SkipVerifyTLS(addr, from, password string, to, cc, bcc []string, sub string, msg []byte) error {
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

	mailfrom, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}
	//auth password
	if password != "" {
		a := smtp.PlainAuth("", mailfrom.Address, password, host)
		if err := c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(mailfrom.Address); err != nil {
		return err
	}

	rcptto, err := parseAddress(to)
	if err != nil {
		return err
	}
	ccto, err := parseAddress(cc)
	if err != nil {
		return err
	}
	bccto, err := parseAddress(bcc)
	if err != nil {
		return err
	}

	for _, rcpt := range rcptto {
		if err := c.Rcpt(rcpt.Address); err != nil {
			return err
		}
	}
	for _, rcpt := range ccto {
		if err := c.Rcpt(rcpt.Address); err != nil {
			return err
		}
	}
	for _, rcpt := range bccto {
		if err := c.Rcpt(rcpt.Address); err != nil {
			return err
		}
	}

	content, err := newMessage(mailfrom, rcptto, ccto, bccto, sub, msg)
	if err != nil {
		return err
	}
	//wc is io.WriteCloser
	wc, err := c.Data()
	if err != nil {
		return err
	}

	// write email content to wc
	if _, err := wc.Write([]byte(content)); err != nil {
		return err
	}
	if err = wc.Close(); err != nil {
		return err
	}
	return c.Quit()
}
