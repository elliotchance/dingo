package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"go/ast"
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

func TestService_ContainerFieldType(t *testing.T) {
	for testName, test := range map[string]struct {
		services           Services
		containerFieldType ast.Expr
	}{
		"StructWithoutScope": {
			services: Services{
				"A": {
					Scope: ScopeNotSet,
					Type:  "*SendEmail",
				},
			},

			// SendEmail is already a pointer, so it will be nil until
			// initialised.
			containerFieldType: newIdent("*SendEmail"),
		},
		"StructContainer": {
			services: Services{
				"A": {
					Scope: ScopeContainer,
					Type:  "SendEmail",
				},
			},

			// SendEmail is not a pointer, so we need to make it one so we know
			// when it is initialised.
			containerFieldType: newIdent("*SendEmail"),
		},
		"StructPrototype": {
			services: Services{
				"A": {
					Scope: ScopePrototype,
					Type:  "*foo.Bar",
				},
			},

			// It must be wrapped in a function so it is created each time.
			containerFieldType: &ast.FuncType{
				Params:  newFieldList(),
				Results: newFieldList("*foo.Bar"),
			},
		},
		"StructContainerWithArguments": {
			services: Services{
				"A": {
					Scope: ScopeContainer,
					Type:  "SendEmail",
					Arguments: Arguments{
						"foo": "int",
					},
				},
			},

			// SendEmail is not a pointer, so we need to make it one so we know
			// when it is initialised.
			containerFieldType: &ast.FuncType{
				Params:  newFieldList("foo int"),
				Results: newFieldList("SendEmail"),
			},
		},
		"StructPrototypeWithArguments": {
			services: Services{
				"A": {
					Scope: ScopePrototype,
					Type:  "*foo.Bar",
					Arguments: Arguments{
						"foo": "int",
						"bar": "float64",
					},
				},
			},

			// It must be wrapped in a function so it is created each time.
			containerFieldType: &ast.FuncType{
				Params:  newFieldList("bar float64, foo int"),
				Results: newFieldList("*foo.Bar"),
			},
		},
		"StructPrototypeWithArgumentsAndDeps1": {
			services: Services{
				"A": {
					Scope:   ScopePrototype,
					Type:    "*foo.Bar",
					Returns: "@{B}",
					Arguments: Arguments{
						"foo": "int",
						"bar": "float64",
					},
				},
				"B": {
					Scope: ScopeContainer,
					Type:  "foo.Baz",
				},
			},
			containerFieldType: &ast.FuncType{
				Params:  newFieldList("B foo.Baz, bar float64, foo int"),
				Results: newFieldList("*foo.Bar"),
			},
		},
		"StructPrototypeWithArgumentsAndDeps2": {
			services: Services{
				"A": {
					Scope:   ScopePrototype,
					Type:    "*foo.Bar",
					Returns: "@{B}",
					Arguments: Arguments{
						"foo": "int",
						"bar": "float64",
					},
				},
				"B": {
					Scope: ScopePrototype,
					Type:  "foo.Baz",
				},
			},

			containerFieldType: &ast.FuncType{
				Params:  newFieldList("B foo.Baz, bar float64, foo int"),
				Results: newFieldList("*foo.Bar"),
			},
		},
		"StructPrototypeWithArgumentsAndDeps3": {
			services: Services{
				"A": {
					Scope:   ScopePrototype,
					Type:    "*foo.Bar",
					Returns: "@{B}",
					Arguments: Arguments{
						"foo": "int",
						"bar": "float64",
					},
				},
				"B": {
					Scope:     ScopePrototype,
					Interface: "Bazer",
					Arguments: Arguments{
						"baz": "time.Time",
					},
				},
			},

			containerFieldType: &ast.FuncType{
				Params:  newFieldList("B Bazer, bar float64, foo int"),
				Results: newFieldList("*foo.Bar"),
			},
		},
		"InterfaceWithoutScope": {
			services: Services{
				"A": {
					Scope:     ScopeNotSet,
					Interface: "Emailer",
				},
			},

			// Interfaces can be nil, so no need to turn it into a pointer.
			containerFieldType: newIdent("Emailer"),
		},
		"InterfaceContainer": {
			services: Services{
				"A": {
					Scope:     ScopeContainer,
					Interface: "Emailer",
				},
			},

			// Interfaces can be nil, so no need to turn it into a pointer.
			containerFieldType: newIdent("Emailer"),
		},
		"InterfacePrototype": {
			services: Services{
				"A": {
					Scope:     ScopePrototype,
					Interface: "Emailer",
				},
			},

			// A func that returns the interface.
			containerFieldType: &ast.FuncType{
				Params:  newFieldList(),
				Results: newFieldList("Emailer"),
			},
		},
		"InterfaceContainerWithArguments": {
			services: Services{
				"A": {
					Scope:     ScopeContainer,
					Interface: "Emailer",
					Arguments: Arguments{
						"foo": "int",
					},
				},
			},

			// Interfaces can be nil, so no need to turn it into a pointer.
			containerFieldType: &ast.FuncType{
				Params:  newFieldList("foo int"),
				Results: newFieldList("Emailer"),
			},
		},
		"InterfacePrototypeWithArguments": {
			services: Services{
				"A": {
					Scope:     ScopePrototype,
					Interface: "Emailer",
					Arguments: Arguments{
						"foo": "int",
						"bar": "float64",
					},
				},
			},

			// A func that returns the interface.
			containerFieldType: &ast.FuncType{
				Params:  newFieldList("bar float64, foo int"),
				Results: newFieldList("Emailer"),
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			actual := test.services["A"].ContainerFieldType(test.services)
			assert.Equal(t, test.containerFieldType, actual)
		})
	}
}
