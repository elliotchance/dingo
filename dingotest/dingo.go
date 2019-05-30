package dingotest

import (
	go_sub_pkg "github.com/elliotchance/dingo/dingotest/go-sub-pkg"
	"os"
)

type Container struct {
	CustomerWelcome	*CustomerWelcome
	OtherPkg	*go_sub_pkg.Person
	OtherPkg2	go_sub_pkg.Greeter
	OtherPkg3	*go_sub_pkg.Person
	SendEmail	EmailSender
	SendEmailError	*SendEmail
	SomeEnv		*string
	WithEnv1	*SendEmail
	WithEnv2	*SendEmail
}

var DefaultContainer = &Container{}

func (container *Container) GetCustomerWelcome() *CustomerWelcome {
	if container.CustomerWelcome == nil {
		service := NewCustomerWelcome(container.GetSendEmail())
		container.CustomerWelcome = service
	}
	return container.CustomerWelcome
}
func (container *Container) GetOtherPkg() *go_sub_pkg.Person {
	if container.OtherPkg == nil {
		service := &go_sub_pkg.Person{}
		container.OtherPkg = service
	}
	return container.OtherPkg
}
func (container *Container) GetOtherPkg2() go_sub_pkg.Greeter {
	if container.OtherPkg2 == nil {
		service := go_sub_pkg.NewPerson()
		container.OtherPkg2 = service
	}
	return container.OtherPkg2
}
func (container *Container) GetOtherPkg3() go_sub_pkg.Person {
	if container.OtherPkg3 == nil {
		service := go_sub_pkg.Person{}
		container.OtherPkg3 = &service
	}
	return *container.OtherPkg3
}
func (container *Container) GetSendEmail() EmailSender {
	if container.SendEmail == nil {
		service := &SendEmail{}
		service.From = "hi@welcome.com"
		container.SendEmail = service
	}
	return container.SendEmail
}
func (container *Container) GetSendEmailError() *SendEmail {
	if container.SendEmailError == nil {
		service, err := NewSendEmail()
		if err != nil {
			panic(err)
		}
		container.SendEmailError = service
	}
	return container.SendEmailError
}
func (container *Container) GetSomeEnv() string {
	if container.SomeEnv == nil {
		service := os.Getenv("ShouldBeSet")
		container.SomeEnv = &service
	}
	return *container.SomeEnv
}
func (container *Container) GetWithEnv1() SendEmail {
	if container.WithEnv1 == nil {
		service := SendEmail{}
		service.From = os.Getenv("ShouldBeSet")
		container.WithEnv1 = &service
	}
	return *container.WithEnv1
}
func (container *Container) GetWithEnv2() *SendEmail {
	if container.WithEnv2 == nil {
		service := &SendEmail{}
		service.From = "foo-" + os.Getenv("ShouldBeSet") + "-bar"
		container.WithEnv2 = service
	}
	return container.WithEnv2
}
