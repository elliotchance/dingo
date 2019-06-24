package main

import (
	"fmt"
	"sort"
)

type Arguments map[string]Type

// Names returns all of the argument names sorted.
func (args Arguments) Names() (names []string) {
	for arg := range args {
		names = append(names, arg)
	}

	sort.Strings(names)

	return
}

func (args Arguments) GoArguments() (ss []string) {
	for _, argName := range args.Names() {
		ss = append(ss, fmt.Sprintf("%s %s", argName,
			args[argName].LocalEntityType()))
	}

	return
}
