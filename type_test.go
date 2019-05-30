package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var typeTests = map[Type]struct {
	IsPointer              bool
	PackageName            string
	LocalPackageName       string
	EntityName             string
	LocalEntityName        string
	LocalEntityType        string
	CreateLocalEntityType  string
	LocalEntityPointerType string
	UnversionedPackageName string
}{
	"Person": {
		IsPointer:              false,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "Person",
		LocalEntityName:        "Person",
		LocalEntityType:        "Person",
		CreateLocalEntityType:  "Person",
		LocalEntityPointerType: "*Person",
		UnversionedPackageName: "",
	},
	"*Person": {
		IsPointer:              true,
		PackageName:            "",
		LocalPackageName:       "",
		EntityName:             "Person",
		LocalEntityName:        "Person",
		LocalEntityType:        "*Person",
		CreateLocalEntityType:  "&Person",
		LocalEntityPointerType: "*Person",
		UnversionedPackageName: "",
	},
	"github.com/elliotchance/dingo/dingotest/go-sub-pkg.Person": {
		IsPointer:              false,
		PackageName:            "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
		LocalPackageName:       "go_sub_pkg",
		EntityName:             "Person",
		LocalEntityName:        "go_sub_pkg.Person",
		LocalEntityType:        "go_sub_pkg.Person",
		CreateLocalEntityType:  "go_sub_pkg.Person",
		LocalEntityPointerType: "*go_sub_pkg.Person",
		UnversionedPackageName: "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
	},
	"*github.com/elliotchance/dingo/dingotest/go-sub-pkg.Person": {
		IsPointer:              true,
		PackageName:            "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
		LocalPackageName:       "go_sub_pkg",
		EntityName:             "Person",
		LocalEntityName:        "go_sub_pkg.Person",
		LocalEntityType:        "*go_sub_pkg.Person",
		CreateLocalEntityType:  "&go_sub_pkg.Person",
		LocalEntityPointerType: "*go_sub_pkg.Person",
		UnversionedPackageName: "github.com/elliotchance/dingo/dingotest/go-sub-pkg",
	},
	"github.com/kounta/luigi/v7.Logger": {
		IsPointer:              false,
		PackageName:            "github.com/kounta/luigi/v7",
		LocalPackageName:       "luigi",
		EntityName:             "Logger",
		LocalEntityName:        "luigi.Logger",
		LocalEntityType:        "luigi.Logger",
		CreateLocalEntityType:  "luigi.Logger",
		LocalEntityPointerType: "*luigi.Logger",
		UnversionedPackageName: "github.com/kounta/luigi",
	},
	"*github.com/kounta/luigi/v7.SimpleLogger": {
		IsPointer:              true,
		PackageName:            "github.com/kounta/luigi/v7",
		LocalPackageName:       "luigi",
		EntityName:             "SimpleLogger",
		LocalEntityName:        "luigi.SimpleLogger",
		LocalEntityType:        "*luigi.SimpleLogger",
		CreateLocalEntityType:  "&luigi.SimpleLogger",
		LocalEntityPointerType: "*luigi.SimpleLogger",
		UnversionedPackageName: "github.com/kounta/luigi",
	},
}

func TestType_String(t *testing.T) {
	for ty := range typeTests {
		t.Run(string(ty), func(t *testing.T) {
			assert.Equal(t, string(ty), ty.String())
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
