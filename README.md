# 🐺 dingo

Easy, fast and type-safe dependency injection for Go.

  * [Installation](#installation)
  * [Building the Container](#building-the-container)
  * [Configuring Services](#configuring-services)
    + [error](#error)
    + [import](#import)
    + [interface](#interface)
    + [properties](#properties)
    + [returns](#returns)
    + [scope](#scope)
    + [type](#type)
  * [Using Services](#using-services)
  * [Unit Testing](#unit-testing)
  * [Practical Examples](#practical-examples)
    + [Mocking the Clock](#mocking-the-clock)
    + [Mocking Runtime Dependencies](#mocking-runtime-dependencies)

## Installation

```bash
go get -u github.com/elliotchance/dingo
```

## Building the Container

Building or rebuilding the container is done with:

```bash
dingo
```

The container is created from a file called `dingo.yml` in the same directory as
where the `dingo` command is run. This should be the root of your
module/repository.

Here is an example of a `dingo.yml`:

```yml
services:
  SendEmail:
    type: '*SendEmail'
    interface: EmailSender
    properties:
      From: '"hi@welcome.com"'

  CustomerWelcome:
    type: '*CustomerWelcome'
    returns: NewCustomerWelcome(@{SendEmail})
```

It will generate a file called `dingo.go`. This must be committed with your
code.

## Configuring Services

The root level `services` key describes each of the services.

The name of the service follows the same naming conventions as Go, so service
names that start with a capital letter will be exported (available outside this
package).

All options described below are optional. However, you must provide either
`type` or `interface`.

Any option below that expects an expression can contain any valid Go code.
References to other services and variables will be substituted automatically:

- `@{SendEmail}` will inject the service named `SendEmail`.
- `${DB_PASS}` will inject the environment variable `DB_PASS`.

### error

If `returns` provides two arguments (where the second one is the error) you must
include an `error`. This is the expression when `err != nil`.

Examples:

- `error: panic(err)` - panic if an error occurs.
- `error: return nil` - return a nil service if an error occurs.

### import

You can provide explicit imports if you need to reference packages in
expressions (such as `returns`) that do not exist in `type` or `interface`.

If a package listed in `import` is already imported, either directly or
indirectly, it value will be ignored.

Example:

```yml
import:
  - 'github.com/aws/aws-sdk-go/aws/session'
```

### interface

If you need to replace this service with another `struct` type in unit tests you
will need to provide an `interface`. This will override `type` and must be
compatible with returned type of `returns`.

Examples:

- `interface: EmailSender` - `EmailSender` in this package.
- `interface: io.Writer` - `Writer` in the `io` package.

### properties

If provided, a map of case-sensitive properties to be set on the instance. Each
of the properties is a Go expression.

Example:

```yml
properties:
  From: "hi@welcome.com"
  maxRetries: 10
  emailer: '@{Emailer}'
```

### returns

The expression used to instantiate the service. You can provide any Go
expression here, including referencing other services and environment variables.

The `returns` can also return a function, since it is an expression. See `type`
for an example.

### scope

The `scope` defines when a service should be created, or when it can be reused.
It must be one of the following values:

- `prototype`: A new instance will be created whenever the service is requested
or injected into another service as a dependency.

- `container` (default): The instance will created once for this container, and
then it will be returned in future requests. This is sometimes called a
singleton, however the service will not be shared outside of the container.

### type

The type returned by the `return` expression. You must provide a fully qualified
name that includes the package name if the type does not belong to this package.

Example

```yml
type: '*github.com/go-redis/redis.Options'
```

The `type` may also be a function. Functions can refer to other services in the
same embedded way:

```yml
type: func () bool
returns: |
  func () bool {
    return @{Something}.IsReady()
  }
```

## Using Services

As part of the generated file, `dingo.go`. There will be a module-level variable
called `DefaultContainer`. This requires no initialization and can be used
immediately:

```go
func main() {
	welcomer := DefaultContainer.GetCustomerWelcome()
	err := welcomer.Welcome("Bob", "bob@smith.com")
	// ...
}
```

## Unit Testing

**When unit testing you should not use the global `DefaultContainer`.** You
should create a new container:

```go
container := NewContainer()
```

Unit tests can make any modifications to the new container, including overriding
services to provide mocks or other stubs:

```go
func TestCustomerWelcome_Welcome(t *testing.T) {
	emailer := FakeEmailSender{}
	emailer.On("Send",
		"bob@smith.com", "Welcome", "Hi, Bob!").Return(nil)
    
	container := NewContainer()
	container.SendEmail = emailer
    
	welcomer := container.GetCustomerWelcome()
	err := welcomer.Welcome("Bob", "bob@smith.com")
	assert.NoError(t, err)
	emailer.AssertExpectations(t)
}
```

## Practical Examples

### Mocking the Clock

Code that relies on time needs to be deterministic to be testable. Extracting
the clock as a service allows the whole time environment to be predictable for
all services. It also has the added benefit that `Sleep()` is free when running
unit tests.

Here is a service, `WhatsTheTime`, that needs to use the current time:

```yml
services:
  Clock:
    interface: github.com/jonboulle/clockwork.Clock
    returns: clockwork.NewRealClock()

  WhatsTheTime:
    type: '*WhatsTheTime'
    properties:
      clock: '@{Clock}'
```

`WhatsTheTime` can now use this clock the same way you would use the `time`
package:

```go
import (
	"github.com/jonboulle/clockwork"
	"time"
)

type WhatsTheTime struct {
	clock clockwork.Clock
}

func (t *WhatsTheTime) InRFC1123() string {
	return t.clock.Now().Format(time.RFC1123)
}
```

The unit test can substitute a fake clock for all services:

```go
func TestWhatsTheTime_InRFC1123(t *testing.T) {
	container := NewContainer()
	container.Clock = clockwork.NewFakeClock()

	actual := container.GetWhatsTheTime().InRFC1123()
	assert.Equal(t, "Wed, 04 Apr 1984 00:00:00 UTC", actual)
}
```

### Mocking Runtime Dependencies

One situation that is tricky to write tests for is when you have the
instantiation inside a service because it needs some runtime state.

Let's say you have a HTTP client that signs a request before sending it. The
signer can only be instantiated with the request, so we can't use traditional
injection:

```go
type HTTPSignerClient struct{}

func (c *HTTPSignerClient) Do(req *http.Request) (*http.Response, error) {
	signer := NewSigner(req)
	req.Headers.Set("Authorization", signer.Auth())

	return http.DefaultClient.Do(req)
}
```

The `Signer` is not deterministic because it relies on the time:

```go
type Signer struct {
	req *http.Request
}

func NewSigner(req *http.Request) *Signer {
	return &Signer{req: req}
}

// Produces something like "Mon Jan 2 15:04:05 2006 POST"
func (signer *Signer) Auth() string {
	return time.Now().Format(time.ANSIC) + " " + signer.req.Method
}
```

Unlike mocking the clock (as in the previous tutorial) this time we need to keep
the logic of the signer, but verify the URL path sent to the signer. Of course,
we could manipulate or entirely replace the signer as well.

Services can have `arguments` which turns them into factories. For example:

```yml
services:
  Signer:
    type: '*Signer'
    scope: prototype        # Create a new Signer each time
    arguments:              # Define the dependencies at runtime.
      req: '*http.Request'
    returns: NewSigner(req) # Setup code can reference the runtime dependencies.

  HTTPSignerClient:
  	type: '*HTTPSignerClient'
  	properties:
  	  CreateSigner: '@{Signer}' # Looks like a regular service, right?
```

Dingo has transformed the service into a factory, using a function:

```go
type HTTPSignerClient struct {
	CreateSigner func(req *http.Request) *Signer
}

func (c *HTTPSignerClient) Do(req *http.Request) (*http.Response, error) {
	signer := c.CreateSigner(req)
	req.Headers.Set("Authorization", signer.Auth())

	return http.DefaultClient.Do(req)
}
```

Under test we can control this factory like any other service:

```go
func TestHTTPSignerClient_Do(t *testing.T) {
	container := NewContainer()
	container.Signer = func(req *http.Request) *Signer {
		assert.Equals(t, req.URL.Path, "/foo")

		return NewSigner(req)
	}

	client := container.GetHTTPSignerClient()
	_, err := client.Do(http.NewRequest("GET", "/foo", nil))
	assert.NoError(t, err)
}
```
