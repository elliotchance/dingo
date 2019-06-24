package main

import (
	"github.com/elliotchance/testify-stats/assert"
	"testing"
)

var argumentTests = map[string]struct {
	Arguments   Arguments
	Names       []string
	GoArguments []string
}{
	"Nil": {
		Arguments:   nil,
		Names:       nil,
		GoArguments: nil,
	},
	"Empty": {
		Arguments:   map[string]Type{},
		Names:       nil,
		GoArguments: nil,
	},
	"One": {
		Arguments:   map[string]Type{"foo": "int"},
		Names:       []string{"foo"},
		GoArguments: []string{"foo int"},
	},
	"ArgumentsAlwaysSortedByName": {
		Arguments:   map[string]Type{"foo": "int", "bar": "*float64"},
		Names:       []string{"bar", "foo"},
		GoArguments: []string{"bar *float64", "foo int"},
	},
	"RemovePackageName": {
		Arguments:   map[string]Type{"req": "*net/http.Request"},
		Names:       []string{"req"},
		GoArguments: []string{"req *http.Request"},
	},
}

func TestArguments_Names(t *testing.T) {
	for testName, test := range argumentTests {
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, test.Names, test.Arguments.Names())
		})
	}
}

func TestArguments_GoArguments(t *testing.T) {
	for testName, test := range argumentTests {
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, test.GoArguments, test.Arguments.GoArguments())
		})
	}
}
