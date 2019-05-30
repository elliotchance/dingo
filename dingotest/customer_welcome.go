package dingotest

import "fmt"

type CustomerWelcome struct {
	Emailer EmailSender
}

func NewCustomerWelcome(sender EmailSender) *CustomerWelcome {
	return &CustomerWelcome{
		Emailer: sender,
	}
}

func (welcomer *CustomerWelcome) Welcome(name, email string) error {
	body := fmt.Sprintf("Hi, %s!", name)
	subject := "Welcome"

	return welcomer.Emailer.Send(email, subject, body)
}
