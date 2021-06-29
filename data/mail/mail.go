package mail

import (
	"github.com/ikeikeikeike/gocore/util"
	"github.com/ikeikeikeike/gocore/util/dsn"
	"github.com/ikeikeikeike/gocore/util/logger"
)

type (
	// Data sends data
	Data struct {
		To      []string
		Bcc     []string
		Cc      []string
		From    string
		Subject string
		Text    []byte
		HTML    []byte
	}

	// Mail provides interface for sends some of kinda E-Mail.
	Mail interface {
		Send(*Data) error
		// TODO: SendWithAttachment
	}
)

func newMail(env util.Environment) Mail {
	mailURI := env.EnvString("MAILURI")

	mdsn, err := dsn.Mail(mailURI)
	if err != nil {
		msg := "[PANIC] failed to parse email uri <%s>: %s"
		logger.Panicf(msg, mailURI, err)
	}

	// smtp or file or stdout
	if mdsn.StdOut {
		msg := "[INFO] A E-Mailer is chosen stdout by <%s>"
		logger.Printf(msg, mailURI)

		return &stdoutMail{dsn: mdsn}

	}

	msg := "[INFO] A E-Mailer is chosen SMTP Server by <%s>"
	logger.Printf(msg, mdsn.Addr)

	return &smtpMail{dsn: mdsn}
}
