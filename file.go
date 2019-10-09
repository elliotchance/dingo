package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type File struct {
	Services Services
	fset     *token.FileSet
	file     *ast.File
}

func ParseYAMLFile(filepath string) (*File, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var all *File
	err = yaml.Unmarshal(f, &all)
	if err != nil {
		return nil, err
	}
	all.fset = token.NewFileSet()
	return all, nil
}

func GenerateContainer(all *File, packageName string, outputFile string) (*File, error) {
	var err error
	packageLine := fmt.Sprintf("// Code generated by dingo; DO NOT EDIT\npackage %s", packageName)
	all.file, err = parser.ParseFile(all.fset, outputFile, packageLine, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	all.file.Decls = append(all.file.Decls,
		all.Services.astContainerStruct(),
		all.Services.astDefaultContainer(),
		all.astNewContainerFunc())

	for _, serviceName := range all.Services.ServiceNames() {
		definition := all.Services[serviceName]

		// Add imports for type, interface and explicit imports.
		for packageName, shortName := range definition.Imports() {
			astutil.AddNamedImport(all.fset, all.file, shortName, packageName)
		}

		all.file.Decls = append(all.file.Decls, &ast.FuncDecl{
			Name: newIdent("Get" + serviceName),
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							newIdent("container"),
						},
						Type: newIdent("*Container"),
					},
				},
			},
			Type: &ast.FuncType{
				Params:  definition.astArguments(),
				Results: newFieldList(definition.InterfaceOrLocalEntityType(all.Services, false)),
			},
			Body: definition.astFunctionBody(all, all.Services, serviceName, serviceName),
		})
	}

	ast.SortImports(all.fset, all.file)

	return all, nil
}

func (file *File) getPackageName(dingoYMLPath string) string {
	abs, err := filepath.Abs(dingoYMLPath)
	if err != nil {
		panic(err)
	}

	// The directory name is not enough because it may contain a command
	// (package main). Find the first non-test file to get the real package
	// name.
	dir := filepath.Dir(abs)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, fileInfo := range files {
		if strings.HasSuffix(fileInfo.Name(), ".go") &&
			!strings.HasSuffix(fileInfo.Name(), "_test.go") {
			f, err := ioutil.ReadFile(dir + "/" + fileInfo.Name())
			if err != nil {
				panic(err)
			}

			parsedFile, err := parser.ParseFile(file.fset, fileInfo.Name(), f, 0)
			if err != nil {
				panic(err)
			}

			return parsedFile.Name.String()
		}
	}

	// Couldn't find the package name. Assume command.
	return "main"
}

func (file *File) astNewContainerFunc() *ast.FuncDecl {
	fields := make(map[string]ast.Expr)

	for _, serviceName := range file.Services.ServicesWithScope(ScopePrototype).ServiceNames() {
		service := file.Services[serviceName]
		fields[serviceName] = &ast.FuncLit{
			Type: service.astFunctionPrototype(file.Services),
			Body: service.astFunctionBody(file, file.Services, "", serviceName),
		}
	}

	return newFunc("NewContainer", nil, []string{"*Container"}, newBlock(
		newReturn(newCompositeLit("&Container", fields)),
	))
}
