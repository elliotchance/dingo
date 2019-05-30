package dingotest

type SendEmail struct {
	From string
}

func (sender *SendEmail) Send(to, subject, body string) error {
	return nil
}

func NewSendEmail() (*SendEmail, error) {
	return &SendEmail{}, nil
}
