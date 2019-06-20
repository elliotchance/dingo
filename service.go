package main

import (
	"fmt"
	"sort"
)

const (
	ScopeNotSet    = ""
	ScopePrototype = "prototype"
	ScopeContainer = "container"
)

type Service struct {
	Error      string
	Import     []string
	Interface  Type
	Properties map[string]string
	Returns    string
	Scope      string
	Type       Type
}

func (service *Service) InterfaceOrLocalEntityType() string {
	if service.Interface != "" {
		return service.Interface.LocalEntityType()
	}

	return service.Type.LocalEntityType()
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
