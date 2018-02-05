package sendmail

import (
	"testing"
)

var (
	sub = []byte("Test sendmail")
	msg = []byte(`<h1>TEST</h1><pre>This email test the action success or failed.
Do not reply it.
</pre>`)
)

//repace the address of email::::
//TestSendmailSkipVerifyTLS test sendmail not verfiy TLS
func TestSendmailSkipVerifyTLS(t *testing.T) {
	var (
		addr = "127.0.0.1:25"
		from = "jnit"
		to   = []string{"wuqingtao@"}
		cc   = []string{"wuqingtao@"}
		bcc  = []string{
			"wuqingtao@",
		}
	)
	err := SendmailSkipVerifyTLS(addr, from, "", to, cc, bcc, sub, msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Action done.\nEmail context:\nAddr: %s;\nFrom: %s;\nTo: %s;\nCc: %s;\nBcc: %s;\nSub: %s;\nMsg: %s;\n",
		addr, from, to, cc, bcc, sub, msg)
}

//TestSendmail test sendmail use smtp.Sendmail
func TestSendmail(t *testing.T) {
	var (
		addr     = "smtp.163.com:25"
		from     = "wqt_1abc2c3z@163.com"
		password = "q"
		to       = []string{"wqt_1abc2c3z@qq.com"}
		cc       = []string{"271abc2c3z9@qq.com"}
		bcc      = []string{
			"wuqingtao@",
		}
	)
	msg = append(msg, []byte("\nno local\n")...)
	err := Sendmail(addr, from, password, to, cc, bcc, sub, msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Action done.\nEmail context:\nAddr: %s;\nFrom: %s;\nTo: %s;\nCc: %s;\nBcc: %s;\nSub: %s;\nMsg: %s;\n",
		addr, from, to, cc, bcc, sub, msg)
}
