package main

import (
	"go/printer"
	"log"
	"os"
	"regexp"
)

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

	file, err := ParseYAMLFile(dingoYMLPath)
	if err != nil {
		log.Fatalln(err)
	}

	packageName := file.Package
	if packageName == "" {
		packageName = file.getPackageName(dingoYMLPath)
	}

	file, err = GenerateContainer(file, packageName, outputFile)
	if err != nil {
		log.Fatalln(err)
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalln("open file:", err)
	}

	err = printer.Fprint(outFile, file.fset, file.file)
	if err != nil {
		log.Fatalln("writer:", err)
	}
}
