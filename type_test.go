package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var typeTests = map[Type]struct {
	String                 string
	IsPointer              bool
	PackageName            string
	LocalPackageName       string
	EntityName             string
	LocalEntityName        string
	LocalEntityType        string
	CreateLocalEntityType  string
	LocalEntityPointerType string
	UnversionedPackageName string
	IsFunction             bool
}{
	"Person": {
		String:                 "Person",
		IsPointer:              false,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "Person",
		LocalEntityName:        "Person",
		LocalEntityType:        "Person",
		CreateLocalEntityType:  "Person",
		LocalEntityPointerType: "*Person",
		UnversionedPackageName: "",
		IsFunction:             false,
	},
	"*Person": {
		String:                 "*Person",
		IsPointer:              true,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "Person",
		LocalEntityName:        "Person",
		LocalEntityType:        "*Person",
		CreateLocalEntityType:  "&Person",
		LocalEntityPointerType: "*Person",
		UnversionedPackageName: "",
		IsFunction:             false,
	},
	"github.com/elliotchance/dingo/dingotest/go-sub-pkg.Person": {
		String:                 "github.com/elliotchance/dingo/dingotest/go-sub-pkg.Person",
		IsPointer:              false,
		PackageName:            "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
		LocalPackageName:       "go_sub_pkg",
		EntityName:             "Person",
		LocalEntityName:        "go_sub_pkg.Person",
		LocalEntityType:        "go_sub_pkg.Person",
		CreateLocalEntityType:  "go_sub_pkg.Person",
		LocalEntityPointerType: "*go_sub_pkg.Person",
		UnversionedPackageName: "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
		IsFunction:             false,
	},
	"*github.com/elliotchance/dingo/dingotest/go-sub-pkg.Person": {
		String:                 "*github.com/elliotchance/dingo/dingotest/go-sub-pkg.Person",
		IsPointer:              true,
		PackageName:            "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
		LocalPackageName:       "go_sub_pkg",
		EntityName:             "Person",
		LocalEntityName:        "go_sub_pkg.Person",
		LocalEntityType:        "*go_sub_pkg.Person",
		CreateLocalEntityType:  "&go_sub_pkg.Person",
		LocalEntityPointerType: "*go_sub_pkg.Person",
		UnversionedPackageName: "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
		IsFunction:             false,
	},
	"github.com/kounta/luigi/v7.Logger": {
		String:                 "github.com/kounta/luigi/v7.Logger",
		IsPointer:              false,
		PackageName:            "github.com/kounta/luigi/v7",
		LocalPackageName:       "luigi",
		EntityName:             "Logger",
		LocalEntityName:        "luigi.Logger",
		LocalEntityType:        "luigi.Logger",
		CreateLocalEntityType:  "luigi.Logger",
		LocalEntityPointerType: "*luigi.Logger",
		UnversionedPackageName: "github.com/kounta/luigi",
		IsFunction:             false,
	},
	"*github.com/kounta/luigi/v7.SimpleLogger": {
		String:                 "*github.com/kounta/luigi/v7.SimpleLogger",
		IsPointer:              true,
		PackageName:            "github.com/kounta/luigi/v7",
		LocalPackageName:       "luigi",
		EntityName:             "SimpleLogger",
		LocalEntityName:        "luigi.SimpleLogger",
		LocalEntityType:        "*luigi.SimpleLogger",
		CreateLocalEntityType:  "&luigi.SimpleLogger",
		LocalEntityPointerType: "*luigi.SimpleLogger",
		UnversionedPackageName: "github.com/kounta/luigi",
		IsFunction:             false,
	},
	"func()": {
		String:                 "func ()",
		IsPointer:              true,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "func ()",
		LocalEntityName:        "func ()",
		LocalEntityType:        "func ()",
		CreateLocalEntityType:  "func ()",
		LocalEntityPointerType: "func ()",
		UnversionedPackageName: "",
		IsFunction:             true,
	},
	"func  (a, b int) (*foo.Bar)": {
		String:                 "func (a, b int) *foo.Bar",
		IsPointer:              true,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "func (a, b int) *foo.Bar",
		LocalEntityName:        "func (a, b int) *foo.Bar",
		LocalEntityType:        "func (a, b int) *foo.Bar",
		CreateLocalEntityType:  "func (a, b int) *foo.Bar",
		LocalEntityPointerType: "func (a, b int) *foo.Bar",
		UnversionedPackageName: "",
		IsFunction:             true,
	},
	"func(s string) (float64, bool)": {
		String:                 "func (s string) (float64, bool)",
		IsPointer:              true,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "func (s string) (float64, bool)",
		LocalEntityName:        "func (s string) (float64, bool)",
		LocalEntityType:        "func (s string) (float64, bool)",
		CreateLocalEntityType:  "func (s string) (float64, bool)",
		LocalEntityPointerType: "func (s string) (float64, bool)",
		UnversionedPackageName: "",
		IsFunction:             true,
	},
}

func TestType_String(t *testing.T) {
	for ty := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, ty.String(), ty.String())
		})
	}
}

func TestType_IsPointer(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.IsPointer, ty.IsPointer())
		})
	}
}

func TestType_PackageName(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.PackageName, ty.PackageName())
		})
	}
}

func TestType_LocalPackageName(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.LocalPackageName, ty.LocalPackageName())
		})
	}
}

func TestType_EntityName(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.EntityName, ty.EntityName())
		})
	}
}

func TestType_LocalEntityName(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.LocalEntityName, ty.LocalEntityName())
		})
	}
}

func TestType_LocalEntityType(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.LocalEntityType, ty.LocalEntityType())
		})
	}
}

func TestType_CreateLocalEntityType(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.CreateLocalEntityType, ty.CreateLocalEntityType())
		})
	}
}

func TestType_LocalEntityPointerType(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.LocalEntityPointerType, ty.LocalEntityPointerType())
		})
	}
}

func TestType_UnversionedPackageName(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.UnversionedPackageName, ty.UnversionedPackageName())
		})
	}
}

func TestType_IsFunction(t *testing.T) {
	for ty, test := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, test.IsFunction, ty.IsFunction())
		})
	}
}
