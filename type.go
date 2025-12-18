package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Type string

func (ty Type) String() string {
	if ty.IsFunction() {
		args, returns := ty.parseFunctionType()
		switch len(returns) {
		case 0:
			return fmt.Sprintf("func (%s)", args)

		case 1:
			return fmt.Sprintf("func (%s) %s", args, strings.Join(returns, ","))

		default:
			return fmt.Sprintf("func (%s) (%s)", args,
				strings.Join(returns, ","))
		}
	}

	return string(ty)
}

func (ty Type) IsPointer() bool {
	return strings.HasPrefix(string(ty), "*") || ty.IsFunction()
}

func (ty Type) PackageName() string {
	if ty.IsFunction() || !strings.Contains(string(ty), ".") {
		return ""
	}

	parts := strings.Split(strings.TrimLeft(string(ty), "*"), ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

func (ty Type) UnversionedPackageName() string {
	if ty.IsFunction() {
		return ""
	}

	packageName := strings.Split(ty.PackageName(), "/")
	if regexp.MustCompile(`^v\d+$`).MatchString(packageName[len(packageName)-1]) {
		packageName = packageName[:len(packageName)-1]
	}

	return strings.Join(packageName, "/")
}

func (ty Type) LocalPackageName(importMap ImportMap) string {
	if ty.IsFunction() {
		return ""
	}

	packageName := ty.PackageName()
	if packageName == "" {
		return ""
	}

	if importMap != nil {
		if resolvedName, ok := importMap[packageName]; ok {
			return resolvedName
		}
	}

	pkgNameParts := strings.Split(ty.UnversionedPackageName(), "/")
	lastPart := pkgNameParts[len(pkgNameParts)-1]
	if lastPart == "" {
		lastPart = packageName
	}
	return strings.Replace(lastPart, "-", "_", -1)
}

func (ty Type) EntityName() string {
	if ty.IsFunction() {
		return ty.String()
	}

	parts := strings.Split(string(ty), ".")

	return strings.TrimLeft(parts[len(parts)-1], "*")
}

func (ty Type) LocalEntityName(importMap ImportMap) string {
	if ty.IsFunction() {
		return ty.String()
	}

	name := ty.LocalPackageName(importMap) + "." + ty.EntityName()

	return strings.TrimLeft(name, ".")
}

func (ty Type) LocalEntityType(importMap ImportMap) string {
	if ty.IsFunction() {
		return ty.String()
	}

	name := ty.LocalEntityName(importMap)
	if ty.IsPointer() {
		name = "*" + name
	}

	return name
}

func (ty Type) CreateLocalEntityType(importMap ImportMap) string {
	if ty.IsFunction() {
		return ty.String()
	}

	name := ty.LocalEntityName(importMap)
	if ty.IsPointer() {
		name = "&" + name
	}

	return name
}

func (ty Type) LocalEntityPointerType(importMap ImportMap) string {
	if ty.IsFunction() {
		return ty.String()
	}

	name := ty.LocalEntityName(importMap)
	if !strings.HasPrefix(name, "*") {
		name = "*" + name
	}

	return name
}

// IsFunction returns true if the type represents a function pointer, like
// "func ()".
func (ty Type) IsFunction() bool {
	return strings.HasPrefix(string(ty), "func")
}

var (
	functionRegexp1 = regexp.MustCompile(`func\s*\((.*?)\)\s*\((.*)\)`)
	functionRegexp2 = regexp.MustCompile(`func\s*\((.*?)\)\s*(.*)`)
)

func (ty Type) splitArgs(s string) []string {
	if s == "" {
		return nil
	}

	return strings.Split(s, ",")
}

func (ty Type) parseFunctionType() (string, []string) {
	matches := functionRegexp1.FindStringSubmatch(string(ty))
	if len(matches) > 0 {
		return matches[1], ty.splitArgs(matches[2])
	}

	matches = functionRegexp2.FindStringSubmatch(string(ty))

	return matches[1], ty.splitArgs(matches[2])
}
