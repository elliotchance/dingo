package main

type Service struct {
	Type       Type
	Interface  Type
	Properties map[string]string
	Returns    string
	Error      string
	Import     []string
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
