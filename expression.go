package main

import (
	"fmt"
	"github.com/elliotchance/pie/pie"
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
	"regexp"
	"strings"
)

type Expression string

func (e Expression) DependencyNames() (deps []string) {
	for _, v := range regexp.MustCompile(`@{(.*?)}`).FindAllStringSubmatch(string(e), -1) {
		parts := strings.Split(v[1], "(")
		deps = append(deps, parts[0])
	}

	return pie.Strings(deps).Unique()
}

func (e Expression) Dependencies() (deps []string) {
	for _, v := range regexp.MustCompile(`@{(.*?)}`).FindAllStringSubmatch(string(e), -1) {
		deps = append(deps, v[1])
	}

	return pie.Strings(deps).Unique()
}

func (e Expression) performSubstitutions(file *File, services Services, fromArgs bool) string {
	stmt := string(e)

	// Replace environment variables.
	stmt = replaceAllStringSubmatchFunc(
		regexp.MustCompile(`\${(.*?)}`), stmt, func(i []string) string {
			astutil.AddImport(file.fset, file.file, "os")

			return fmt.Sprintf("os.Getenv(\"%s\")", i[1])
		})

	// Replace service names.
	stmt = replaceAllStringSubmatchFunc(
		regexp.MustCompile(`@{(.*?)}`), stmt, func(i []string) string {
			if fromArgs {
				return strings.Split(i[1], "(")[0]
			}

			if strings.Contains(i[1], "(") {
				return fmt.Sprintf("container.Get%s", i[1])
			}

			if _, ok := services[i[1]].ContainerFieldType(services).(*ast.FuncType); ok {
				return fmt.Sprintf("container.%s", i[1])
			}

			return fmt.Sprintf("container.Get%s()", i[1])
		})

	return stmt
}
