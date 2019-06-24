package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

var serviceValidationTests = map[string]struct {
	service *Service
	err     error
}{
	"empty": {
		service: &Service{},
		err:     nil,
	},
	"scope_prototype": {
		service: &Service{
			Scope: "prototype",
		},
		err: nil,
	},
	"scope_container": {
		service: &Service{
			Scope: "container",
		},
		err: nil,
	},
	"scope_invalid": {
		service: &Service{
			Scope: "foo",
		},
		err: errors.New("invalid scope: foo"),
	},
}

func TestService_Validate(t *testing.T) {
	for testName, test := range serviceValidationTests {
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, test.err, test.service.Validate())
		})
	}
}
