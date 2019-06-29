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

func (service *Service) ContainerFieldType(services Services) ast.Expr {
	scope := service.Scope
	if scope == ScopeNotSet {
		scope = ScopeContainer
	}

	if scope == ScopeContainer && len(service.Arguments) == 0 {
		return newIdent(service.InterfaceOrLocalEntityPointerType())
	}

	return service.astFunctionPrototype(services)
}

func (service *Service) InterfaceOrLocalEntityType(services Services, recurse bool) string {
	localEntityType := service.Type.LocalEntityType()
	if service.Interface != "" {
		localEntityType = service.Interface.LocalEntityType()
	}

	if len(service.Arguments) > 0 && recurse {
		var args []string

		for _, dep := range service.Returns.Dependencies() {
			ty := services[dep].InterfaceOrLocalEntityType(services, false)
			args = append(args, fmt.Sprintf("%s %s", dep, ty))
		}

		args = append(args, service.Arguments.GoArguments()...)

		return fmt.Sprintf("func(%v) %s", strings.Join(args, ", "),
			localEntityType)
	}

	return localEntityType
}

func (service *Service) InterfaceOrLocalEntityPointerType() string {
	if service.Interface != "" {
		return service.Interface.LocalEntityType()
	}

	return service.Type.LocalEntityPointerType()
}

func (service *Service) Imports() map[string]string {
	imports := map[string]string{}

	for _, packageName := range service.Import {
		imports[packageName] = ""
	}

	if service.Type.PackageName() != "" {
		imports[service.Type.PackageName()] = service.Type.LocalPackageName()
	}

	if service.Interface.PackageName() != "" {
		imports[service.Interface.PackageName()] = service.Interface.LocalPackageName()
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

	for arg, ty := range service.Arguments {
		funcParams.List = append(funcParams.List, &ast.Field{
			Type: &ast.Ident{
				Name: string(arg + " " + ty.String()),
			},
		})
	}

	return funcParams
}

func (service *Service) astDependencyArguments(services Services) *ast.FieldList {
	funcParams := &ast.FieldList{
		List: []*ast.Field{},
	}

	for _, dep := range service.Returns.DependencyNames() {
		funcParams.List = append(funcParams.List, &ast.Field{
			Type: newIdent(dep + " " + services[dep].InterfaceOrLocalEntityType(services, false)),
		})
	}

	return funcParams
}

func (service *Service) astAllArguments(services Services) *ast.FieldList {
	deps := service.astDependencyArguments(services)
	args := service.astArguments()

	return &ast.FieldList{
		List: append(deps.List, args.List...),
	}
}

func (service *Service) astFunctionPrototype(services Services) *ast.FuncType {
	ty := Type(service.InterfaceOrLocalEntityType(services, true))
	if ty.IsFunction() {
		args, returns := ty.parseFunctionType()

		return &ast.FuncType{
			Params:  newFieldList(args),
			Results: newFieldList(returns...),
		}
	}

	return &ast.FuncType{
		Params:  service.astAllArguments(services),
		Results: newFieldList(string(ty)),
	}
}

func (service *Service) astFunctionBody(file *File, services Services, name, serviceName string) *ast.BlockStmt {
	if name != "" && service.Scope == ScopePrototype {
		var arguments []string
		for _, dep := range service.Returns.Dependencies() {
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
						Type: newIdent(service.Type.CreateLocalEntityType()),
					},
				},
			},
		}
	} else {
		lhs := []ast.Expr{newIdent(serviceTempVariable)}

		if service.Error != "" {
			lhs = append(lhs, newIdent("err"))
		}

		instantiation = []ast.Stmt{
			&ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: lhs,
				Rhs: []ast.Expr{
					newIdent(service.Returns.performSubstitutions(file, services, name == "")),
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
		instantiation = append(instantiation, &ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{&ast.Ident{Name: serviceTempVariable + "." + property.Name}},
			Rhs: []ast.Expr{&ast.Ident{Name: property.Value.performSubstitutions(file, services, name == "")}},
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
