package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"unicode"

	"github.com/acim/arc/pkg/mail"
	"github.com/acim/arc/pkg/middleware"
	"github.com/asaskevich/govalidator"
	"go.uber.org/zap"
)

// ErrInvalidValue is returned when a value is not defined or has wrong value.
var ErrInvalidValue = errors.New("invalid value")

// Mail controller.
type Mail struct {
	mail   mail.Sender
	to     string
	logger *zap.Logger
}

// NewMail creates new mail controller.
func NewMail(sender mail.Sender, recipient string, logger *zap.Logger) *Mail {
	return &Mail{
		mail:   sender,
		to:     recipient,
		logger: logger,
	}
}

// Send sends e-mail from contact form.
func (c *Mail) Send(w http.ResponseWriter, r *http.Request) {
	res := middleware.ResponseFromContext(r.Context())

	mr := &mailReq{} //nolint:exhaustivestruct

	err := json.NewDecoder(r.Body).Decode(mr)
	if err != nil {
		c.logger.Warn("send", zap.NamedError("json decode", err))
		res.SetStatusBadRequest(errParsingRequestBody)

		return
	}

	if err = mr.validate(); err != nil {
		c.logger.Warn("send", zap.NamedError("validate", err))
		res.SetStatusBadRequest(firstToUpper(err.Error()))

		return
	}

	mres, err := c.mail.Send(r.Context(), &mail.Mail{
		From:    fmt.Sprintf("%s %s %s <%s>", mr.FirstName, mr.LastName, mr.Company, mr.From),
		Subject: mr.Subject,
		Text:    mr.Text,
		To:      []string{c.to},
	})
	if err != nil {
		c.logger.Error("send", zap.Error(err), zap.String("response", mres.Message), zap.String("mailgun id", mres.ID))
		res.SetStatusInternalServerError("Error sending e-mail")

		return
	}

	res.SetStatusAccepted()
}

type mailReq struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Company   string `json:"company"`
	From      string `json:"from"`
	Subject   string `json:"subject"`
	Text      string `json:"text"`
}

// Validate input data.
func (m *mailReq) validate() error {
	if (m.FirstName == "" && m.LastName == "") || m.From == "" || m.Subject == "" || m.Text == "" {
		return fmt.Errorf("name, e-mail, subject or message: %w", ErrInvalidValue)
	}

	if !govalidator.IsEmail(m.From) {
		return fmt.Errorf("from address: %w", ErrInvalidValue)
	}

	return nil
}

func firstToUpper(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}

	return ""
}
