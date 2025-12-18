package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"
)

const (
	ScopeNotSet    = ""
	ScopePrototype = "prototype"
	ScopeContainer = "container"
)

type Service struct {
	Arguments  Arguments
	Error      string
	Import     []string
	Interface  Type
	Properties map[string]Expression
	Returns    Expression
	Scope      string
	Type       Type
}

func (service *Service) ContainerFieldType(services Services, importMap ImportMap) ast.Expr {
	scope := service.Scope
	if scope == ScopeNotSet {
		scope = ScopeContainer
	}

	if scope == ScopeContainer && len(service.Arguments) == 0 {
		return newIdent(service.InterfaceOrLocalEntityPointerType(importMap))
	}

	return service.astFunctionPrototype(services, importMap)
}

func (service *Service) InterfaceOrLocalEntityType(services Services, recurse bool, importMap ImportMap) string {
	localEntityType := service.Type.LocalEntityType(importMap)
	if service.Interface != "" {
		localEntityType = service.Interface.LocalEntityType(importMap)
	}

	if len(service.Arguments) > 0 && recurse {
		var args []string

		for _, dep := range service.Returns.Dependencies() {
			ty := services[dep].InterfaceOrLocalEntityType(services, false, importMap)
			args = append(args, fmt.Sprintf("%s %s", dep, ty))
		}

		args = append(args, service.Arguments.GoArguments()...)

		return fmt.Sprintf("func(%v) %s", strings.Join(args, ", "),
			localEntityType)
	}

	return localEntityType
}

func (service *Service) InterfaceOrLocalEntityPointerType(importMap ImportMap) string {
	if service.Interface != "" {
		return service.Interface.LocalEntityType(importMap)
	}

	return service.Type.LocalEntityPointerType(importMap)
}

func (service *Service) Imports() map[string]string {
	imports := map[string]string{}

	for _, packageName := range service.Import {
		imports[packageName] = ""
	}

	if service.Type.PackageName() != "" {
		imports[service.Type.PackageName()] = service.Type.LocalPackageName(nil)
	}

	if service.Interface.PackageName() != "" {
		imports[service.Interface.PackageName()] = service.Interface.LocalPackageName(nil)
	}

	return imports
}


func (service *Service) SortedProperties() (sortedProperties []*Property) {
	var propertyNames []string
	for propertyName := range service.Properties {
		propertyNames = append(propertyNames, propertyName)
	}

	sort.Strings(propertyNames)

	for _, propertyName := range propertyNames {
		sortedProperties = append(sortedProperties, &Property{
			Name:  propertyName,
			Value: service.Properties[propertyName],
		})
	}

	return
}

func (service *Service) ValidateScope() error {
	switch service.Scope {
	case ScopeNotSet, ScopePrototype, ScopeContainer:
		return nil
	}

	return fmt.Errorf("invalid scope: %s", service.Scope)
}

func (service *Service) Validate() error {
	if err := service.ValidateScope(); err != nil {
		return err
	}

	return nil
}

func (service *Service) astArguments() *ast.FieldList {
	funcParams := &ast.FieldList{
		List: []*ast.Field{},
	}

	for _, arg := range service.Arguments.Names() {
		funcParams.List = append(funcParams.List, &ast.Field{
			Type: &ast.Ident{
				Name: string(arg + " " + service.Arguments[arg].String()),
			},
		})
	}

	return funcParams
}

func (service *Service) astDependencyArguments(services Services, importMap ImportMap) *ast.FieldList {
	funcParams := &ast.FieldList{
		List: []*ast.Field{},
	}

	for _, dep := range service.Returns.DependencyNames() {
		funcParams.List = append(funcParams.List, &ast.Field{
			Type: newIdent(dep + " " + services[dep].InterfaceOrLocalEntityType(services, false, importMap)),
		})
	}

	return funcParams
}

func (service *Service) astAllArguments(services Services, importMap ImportMap) *ast.FieldList {
	deps := service.astDependencyArguments(services, importMap)
	args := service.astArguments()

	return &ast.FieldList{
		List: append(deps.List, args.List...),
	}
}

func (service *Service) astFunctionPrototype(services Services, importMap ImportMap) *ast.FuncType {
	ty := Type(service.InterfaceOrLocalEntityType(services, true, importMap))
	if ty.IsFunction() {
		args, returns := ty.parseFunctionType()

		return &ast.FuncType{
			Params:  newFieldList(args),
			Results: newFieldList(returns...),
		}
	}

	return &ast.FuncType{
		Params:  service.astAllArguments(services, importMap),
		Results: newFieldList(string(ty)),
	}
}

func (service *Service) astFunctionBody(file *File, services Services, name, serviceName string) *ast.BlockStmt {
	if name != "" && service.Scope == ScopePrototype {
		var arguments []string
		for _, dep := range service.Returns.Dependencies() {
			if dep[len(dep)-1:]!=")" {
				dep = dep+"()"
			}
			arguments = append(arguments, fmt.Sprintf("container.Get%s", dep))
		}
		arguments = append(arguments, service.Arguments.Names()...)

		return newBlock(
			newReturn(newIdent("container." + serviceName + "(" + strings.Join(arguments, ", ") + ")")),
		)
	}

	var stmts, instantiation []ast.Stmt
	serviceVariable := "container." + name
	serviceTempVariable := "service"

	// Instantiation
	if service.Returns == "" {
		instantiation = []ast.Stmt{
			&ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: []ast.Expr{newIdent(serviceTempVariable)},
				Rhs: []ast.Expr{
					&ast.CompositeLit{
						Type: newIdent(service.Type.CreateLocalEntityType(file.importMap)),
					},
				},
			},
		}
	} else {
		lhs := []ast.Expr{newIdent(serviceTempVariable)}

		if service.Error != "" {
			lhs = append(lhs, newIdent("err"))
		}

		// Perform substitutions with context awareness for package prefixes
		substitutedReturns := service.Returns.performSubstitutions(file, services, name == "")
		if file.importMap != nil {
			substitutedReturns = service.Returns.replacePackagePrefixesWithContext(substitutedReturns, service.Type, file.importMap)
		}

		instantiation = []ast.Stmt{
			&ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: lhs,
				Rhs: []ast.Expr{
					newIdent(substitutedReturns),
				},
			},
		}

		if service.Error != "" {
			instantiation = append(instantiation, &ast.IfStmt{
				Cond: newIdent("err != nil"),
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: newIdent(service.Error),
						},
					},
				},
			})
		}
	}

	// Properties
	for _, property := range service.SortedProperties() {
		substitutedValue := property.Value.performSubstitutions(file, services, name == "")
		if file.importMap != nil {
			substitutedValue = property.Value.replacePackagePrefixesWithContext(substitutedValue, service.Type, file.importMap)
		}

		instantiation = append(instantiation, &ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{&ast.Ident{Name: serviceTempVariable + "." + property.Name}},
			Rhs: []ast.Expr{&ast.Ident{Name: substitutedValue}},
		})
	}

	// Scope
	switch service.Scope {
	case ScopeNotSet, ScopeContainer:
		if service.Type.IsPointer() || service.Interface != "" {
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

		stmts = append(stmts, &ast.IfStmt{
			Cond: &ast.Ident{Name: serviceVariable + " == nil"},
			Body: &ast.BlockStmt{
				List: instantiation,
			},
		})

		// Returns
		if service.Type.IsPointer() || service.Interface != "" {
			stmts = append(stmts, newReturn(newIdent(serviceVariable)))
		} else {
			stmts = append(stmts, newReturn(newIdent("*"+serviceVariable)))
		}

	case ScopePrototype:
		stmts = append(stmts, instantiation...)
		stmts = append(stmts, newReturn(newIdent("service")))
	}

	return newBlock(stmts...)
}
