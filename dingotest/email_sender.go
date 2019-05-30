package dingotest

type EmailSender interface {
	Send(to, subject, body string) error
}
