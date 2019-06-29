package main

import (
	"go/ast"
	"go/token"
	"sort"
)

type Services map[string]*Service

func (services Services) ServiceNames() []string {
	// Sort services to the output file is neat and deterministic.
	var serviceNames []string
	for name := range services {
		serviceNames = append(serviceNames, name)
	}

	sort.Strings(serviceNames)

	return serviceNames
}

func (services Services) ServicesWithScope(scope string) Services {
	ss := make(Services)

	for serviceName, service := range services {
		if service.Scope == scope {
			ss[serviceName] = service
		}
	}

	return ss
}

// astContainer creates the Container struct.
func (services Services) astContainerStruct() *ast.GenDecl {
	var containerFields []*ast.Field
	for _, serviceName := range services.ServiceNames() {
		service := services[serviceName]

		containerFields = append(containerFields, &ast.Field{
			Names: []*ast.Ident{
				{Name: serviceName},
			},
			Type: service.ContainerFieldType(services),
		})
	}

	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: newIdent("Container"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: containerFields,
					},
				},
			},
		},
	}
}

func (services Services) astDefaultContainer() *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					{Name: "DefaultContainer"},
				},
				Values: []ast.Expr{
					&ast.Ident{Name: "NewContainer()"},
				},
			},
		},
	}
}
