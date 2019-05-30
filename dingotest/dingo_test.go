package dingotest_test

import (
	"github.com/elliotchance/dingo/dingotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

type FakeEmailSender struct {
	mock.Mock
}

func (mock *FakeEmailSender) Send(to, subject, body string) error {
	args := mock.Called(to, subject, body)
	return args.Error(0)
}

func TestMain(m *testing.M) {
	_ = os.Setenv("ShouldBeSet", "qux")
	os.Exit(m.Run())
}

func TestCustomerWelcome_Welcome(t *testing.T) {
	emailer := &FakeEmailSender{}
	emailer.On("Send",
		"bob@smith.com", "Welcome", "Hi, Bob!").Return(nil)

	container := &dingotest.Container{}
	container.SendEmail = emailer

	welcomer := container.GetCustomerWelcome()
	err := welcomer.Welcome("Bob", "bob@smith.com")
	assert.NoError(t, err)
	emailer.AssertExpectations(t)
}

func TestDefaultContainer(t *testing.T) {
	assert.NotNil(t, dingotest.DefaultContainer)
	assert.Nil(t, dingotest.DefaultContainer.SendEmail)
	assert.Nil(t, dingotest.DefaultContainer.CustomerWelcome)
}

func TestContainer_GetSendEmail(t *testing.T) {
	container := &dingotest.Container{}

	assert.Nil(t, container.SendEmail)

	// Is singleton.
	service1 := container.GetSendEmail()
	assert.IsType(t, (*dingotest.SendEmail)(nil), service1)

	service2 := container.GetSendEmail()
	assert.IsType(t, (*dingotest.SendEmail)(nil), service2)
	assert.Exactly(t, service1, service2)

	// Properties
	assert.Equal(t, "hi@welcome.com", service1.(*dingotest.SendEmail).From)
	assert.Equal(t, "hi@welcome.com", service2.(*dingotest.SendEmail).From)

	assert.NotNil(t, container.SendEmail)
}

func TestContainer_GetCustomerWelcome(t *testing.T) {
	container := &dingotest.Container{}

	assert.Nil(t, container.SendEmail)
	assert.Nil(t, container.CustomerWelcome)

	// Is singleton.
	service1 := container.GetCustomerWelcome()
	assert.IsType(t, (*dingotest.CustomerWelcome)(nil), service1)

	service2 := container.GetCustomerWelcome()
	assert.IsType(t, (*dingotest.CustomerWelcome)(nil), service2)
	assert.Exactly(t, service1, service2)

	assert.NotNil(t, container.SendEmail)
	assert.NotNil(t, container.CustomerWelcome)
}

func TestContainer_GetWithEnv1(t *testing.T) {
	container := &dingotest.Container{}

	service := container.GetWithEnv1()
	assert.Equal(t, "qux", service.From)
}

func TestContainer_GetWithEnv2(t *testing.T) {
	container := &dingotest.Container{}

	service := container.GetWithEnv2()
	assert.Equal(t, "foo-qux-bar", service.From)
}

func TestContainer_GetSomeEnv(t *testing.T) {
	container := &dingotest.Container{}

	service := container.GetSomeEnv()
	assert.Equal(t, "qux", service)
}
