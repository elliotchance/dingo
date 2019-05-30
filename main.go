package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var fset *token.FileSet
var file *ast.File

type File struct {
	Services map[string]Service
}

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

func resolveStatement(stmt string) string {
	// Replace environment variables.
	stmt = replaceAllStringSubmatchFunc(
		regexp.MustCompile(`\${(.*?)}`), stmt, func(i []string) string {
			astutil.AddImport(fset, file, "os")

			return fmt.Sprintf("os.Getenv(\"%s\")", i[1])
		})

	// Replace service names.
	stmt = replaceAllStringSubmatchFunc(
		regexp.MustCompile(`@{(.*?)}`), stmt, func(i []string) string {
			return fmt.Sprintf("container.Get%s()", i[1])
		})

	return stmt
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

	// Sort services to the output file is neat and deterministic.
	var serviceNames []string
	for name := range all.Services {
		serviceNames = append(serviceNames, name)
	}

	sort.Strings(serviceNames)

	// type Container struct
	var containerFields []*ast.Field
	for _, serviceName := range serviceNames {
		definition := all.Services[serviceName]

		containerFields = append(containerFields, &ast.Field{
			Names: []*ast.Ident{
				{Name: serviceName},
			},
			Type: &ast.Ident{Name: definition.InterfaceOrLocalEntityPointerType()},
		})
	}

	file.Decls = append(file.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: "Container"},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: containerFields,
					},
				},
			},
		},
	})

	file.Decls = append(file.Decls, &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					{Name: "DefaultContainer"},
				},
				Values: []ast.Expr{
					&ast.Ident{Name: "&Container{}"},
				},
			},
		},
	})

	for _, serviceName := range serviceNames {
		definition := all.Services[serviceName]

		// Add imports for type, interface and explicit imports.
		for packageName, shortName := range definition.Imports() {
			astutil.AddNamedImport(fset, file, shortName, packageName)
		}

		returnTypeParts := strings.Split(
			regexp.MustCompile(`/v\d+\.`).ReplaceAllString(string(definition.Type), "."), "/")
		returnType := returnTypeParts[len(returnTypeParts)-1]
		if strings.HasPrefix(string(definition.Type), "*") && !strings.HasPrefix(returnType, "*") {
			returnType = "*" + returnType
		}

		var stmts, instantiation []ast.Stmt
		serviceVariable := "container." + serviceName
		serviceTempVariable := "service"

		// Instantiation
		if definition.Returns == "" {
			instantiation = []ast.Stmt{
				&ast.AssignStmt{
					Tok: token.DEFINE,
					Lhs: []ast.Expr{&ast.Ident{Name: serviceTempVariable}},
					Rhs: []ast.Expr{
						&ast.CompositeLit{
							Type: &ast.Ident{Name: definition.Type.CreateLocalEntityType()},
						},
					},
				},
			}
		} else {
			lhs := []ast.Expr{&ast.Ident{Name: serviceTempVariable}}

			if definition.Error != "" {
				lhs = append(lhs, &ast.Ident{Name: "err"})
			}

			instantiation = []ast.Stmt{
				&ast.AssignStmt{
					Tok: token.DEFINE,
					Lhs: lhs,
					Rhs: []ast.Expr{&ast.Ident{Name: resolveStatement(definition.Returns)}},
				},
			}

			if definition.Error != "" {
				instantiation = append(instantiation, &ast.IfStmt{
					Cond: &ast.Ident{Name: "err != nil"},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.Ident{Name: definition.Error},
							},
						},
					},
				})
			}
		}

		// Properties
		for propertyName, propertyValue := range definition.Properties {
			instantiation = append(instantiation, &ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{&ast.Ident{Name: serviceTempVariable + "." + propertyName}},
				Rhs: []ast.Expr{&ast.Ident{Name: resolveStatement(propertyValue)}},
			})
		}

		if definition.Type.IsPointer() || definition.Interface != "" {
			instantiation = append(instantiation, &ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{&ast.Ident{Name: serviceVariable}},
				Rhs: []ast.Expr{&ast.Ident{Name: serviceTempVariable}},
			})
		} else {
			instantiation = append(instantiation, &ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{&ast.Ident{Name: serviceVariable}},
				Rhs: []ast.Expr{&ast.Ident{Name: "&" + serviceTempVariable}},
			})
		}

		// Singleton
		stmts = append(stmts, &ast.IfStmt{
			Cond: &ast.Ident{Name: serviceVariable + " == nil"},
			Body: &ast.BlockStmt{
				List: instantiation,
			},
		})

		// Return
		if definition.Type.IsPointer() || definition.Interface != "" {
			stmts = append(stmts, &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: serviceVariable},
				},
			})
		} else {
			stmts = append(stmts, &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "*" + serviceVariable},
				},
			})
		}

		file.Decls = append(file.Decls, &ast.FuncDecl{
			Name: &ast.Ident{Name: "Get" + serviceName},
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "container"},
						},
						Type: &ast.Ident{Name: "*Container"},
					},
				},
			},
			Type: &ast.FuncType{
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.Ident{Name: definition.InterfaceOrLocalEntityType()},
						},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: stmts,
			},
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
