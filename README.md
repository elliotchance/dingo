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

The expression used to instantiate the service. You can provide any Go code
here, including referencing other services and environment variables.

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

```:
type: '*github.com/go-redis/redis.Options'
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
container := &Container{}
```

Unit tests can make any modifications to the new container, including overriding
services to provide mocks or other stubs:

```go
func TestCustomerWelcome_Welcome(t *testing.T) {
	emailer := FakeEmailSender{}
	emailer.On("Send",
		"bob@smith.com", "Welcome", "Hi, Bob!").Return(nil)
    
	container := &Container{}
	container.SendEmail = emailer
    
	welcomer := container.GetCustomerWelcome()
	err := welcomer.Welcome("Bob", "bob@smith.com")
	assert.NoError(t, err)
	emailer.AssertExpectations(t)
}
```
