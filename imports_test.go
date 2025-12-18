package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectAndResolveImports_NoConflicts(t *testing.T) {
	services := Services{
		"Service1": &Service{
			Type: "app/foo.Foo",
		},
		"Service2": &Service{
			Type: "app/bar.Bar",
		},
	}

	importMap := CollectAndResolveImports(services)

	assert.Equal(t, "foo", importMap["app/foo"])
	assert.Equal(t, "bar", importMap["app/bar"])
}

func TestCollectAndResolveImports_WithConflicts(t *testing.T) {
	services := Services{
		"HttpClient": &Service{
			Type: "app/http/clients.Client",
		},
		"GrpcClient": &Service{
			Type: "app/grpc/clients.Client",
		},
	}

	importMap := CollectAndResolveImports(services)

	// Both should have resolved names that are different
	httpName := importMap["app/http/clients"]
	grpcName := importMap["app/grpc/clients"]

	assert.NotEqual(t, httpName, grpcName, "Resolved names should be different")
	assert.NotEmpty(t, httpName)
	assert.NotEmpty(t, grpcName)

	// The names should contain distinguishing information
	assert.Contains(t, httpName, "http")
	assert.Contains(t, grpcName, "grpc")
}

func TestCollectAndResolveImports_MultipleConflicts(t *testing.T) {
	services := Services{
		"Service1": &Service{
			Type: "github.com/org1/utils.Helper",
		},
		"Service2": &Service{
			Type: "github.com/org2/utils.Helper",
		},
		"Service3": &Service{
			Type: "github.com/org3/utils.Helper",
		},
	}

	importMap := CollectAndResolveImports(services)

	// All three should have different resolved names
	name1 := importMap["github.com/org1/utils"]
	name2 := importMap["github.com/org2/utils"]
	name3 := importMap["github.com/org3/utils"]

	assert.NotEqual(t, name1, name2)
	assert.NotEqual(t, name2, name3)
	assert.NotEqual(t, name1, name3)

	// Each should contain their distinguishing parent directory
	assert.Contains(t, name1, "org1")
	assert.Contains(t, name2, "org2")
	assert.Contains(t, name3, "org3")
}

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with-dash", "with_dash"},
		{"with.dot", "with_dot"},
		{"with/slash", "with_slash"},
		{"MixedCase", "MixedCase"},
		{"with-multiple-dashes", "with_multiple_dashes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizePackageName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateUniqueName(t *testing.T) {
	tests := []struct {
		name         string
		packagePath  string
		baseShort    string
		allPaths     []string
		shouldContain string
	}{
		{
			name:         "http clients vs grpc clients",
			packagePath:  "app/http/clients",
			baseShort:    "clients",
			allPaths:     []string{"app/http/clients", "app/grpc/clients"},
			shouldContain: "http",
		},
		{
			name:         "grpc clients vs http clients",
			packagePath:  "app/grpc/clients",
			baseShort:    "clients",
			allPaths:     []string{"app/http/clients", "app/grpc/clients"},
			shouldContain: "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUniqueName(tt.packagePath, tt.baseShort, tt.allPaths)
			assert.Contains(t, result, tt.shouldContain)
		})
	}
}
