package main

import (
	"regexp"
	"strings"
)

type Type string

func (ty Type) String() string {
	return string(ty)
}

func (ty Type) IsPointer() bool {
	return strings.HasPrefix(string(ty), "*")
}

func (ty Type) PackageName() string {
	if !strings.Contains(string(ty), ".") {
		return ""
	}

	parts := strings.Split(strings.TrimLeft(string(ty), "*"), ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

func (ty Type) UnversionedPackageName() string {
	packageName := strings.Split(ty.PackageName(), "/")
	if regexp.MustCompile(`^v\d+$`).MatchString(packageName[len(packageName)-1]) {
		packageName = packageName[:len(packageName)-1]
	}

	return strings.Join(packageName, "/")
}

func (ty Type) LocalPackageName() string {
	pkgNameParts := strings.Split(ty.UnversionedPackageName(), "/")
	lastPart := pkgNameParts[len(pkgNameParts)-1]

	return strings.Replace(lastPart, "-", "_", -1)
}

func (ty Type) EntityName() string {
	parts := strings.Split(string(ty), ".")

	return strings.TrimLeft(parts[len(parts)-1], "*")
}

func (ty Type) LocalEntityName() string {
	name := ty.LocalPackageName() + "." + ty.EntityName()

	return strings.TrimLeft(name, ".")
}

func (ty Type) LocalEntityType() string {
	name := ty.LocalEntityName()
	if ty.IsPointer() {
		name = "*" + name
	}

	return name
}

func (ty Type) CreateLocalEntityType() string {
	name := ty.LocalEntityName()
	if ty.IsPointer() {
		name = "&" + name
	}

	return name
}

func (ty Type) LocalEntityPointerType() string {
	name := ty.LocalEntityName()
	if !strings.HasPrefix(name, "*") {
		name = "*" + name
	}

	return name
}
