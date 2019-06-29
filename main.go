package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var fset *token.FileSet
var file *ast.File

func replaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		var groups []string
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func main() {
	dingoYMLPath := "dingo.yml"
	outputFile := "dingo.go"

	f, err := ioutil.ReadFile(dingoYMLPath)
	if err != nil {
		log.Fatalln("reading file:", err)
	}

	all := File{}
	err = yaml.Unmarshal(f, &all)
	if err != nil {
		log.Fatalln("yaml:", err)
	}

	fset = token.NewFileSet()
	packageLine := fmt.Sprintf("package %s", getPackageName(dingoYMLPath))
	file, err = parser.ParseFile(fset, outputFile, packageLine, 0)
	if err != nil {
		log.Fatalln("parser:", err)
	}

	file.Decls = append(file.Decls,
		all.Services.astContainerStruct(),
		all.Services.astDefaultContainer(),
		all.Services.astNewContainerFunc())

	for _, serviceName := range all.Services.ServiceNames() {
		definition := all.Services[serviceName]

		// Add imports for type, interface and explicit imports.
		for packageName, shortName := range definition.Imports() {
			astutil.AddNamedImport(fset, file, shortName, packageName)
		}

		file.Decls = append(file.Decls, &ast.FuncDecl{
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
			Body: definition.astFunctionBody(all.Services, serviceName, serviceName),
		})
	}

	ast.SortImports(fset, file)

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalln("open file:", err)
	}

	err = printer.Fprint(outFile, fset, file)
	if err != nil {
		log.Fatalln("writer:", err)
	}
}

func getPackageName(dingoYMLPath string) string {
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

			file, err = parser.ParseFile(fset, fileInfo.Name(), f, 0)
			if err != nil {
				panic(err)
			}

			return file.Name.String()
		}
	}

	// Couldn't find the package name. Assume command.
	return "main"
}
