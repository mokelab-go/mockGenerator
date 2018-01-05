package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	generator "github.com/mokelab-go/mockGenerator/generator/if"
)

func main() {
	inputFileName := flag.String("i", "", "src file")
	outputFile := flag.String("o", "", "output file")
	flag.Parse()

	if len(*inputFileName) == 0 {
		fmt.Printf("no input file\n")
		return
	}
	if len(*outputFile) == 0 {
		fmt.Printf("no output file\n")
		return
	}

	src, err := ioutil.ReadFile(*inputFileName)
	if err != nil {
		fmt.Printf("Failed to open input file %s", err)
		return
	}

	g := generator.New()
	out, err := g.Generate(string(src))
	if err != nil {
		fmt.Printf("Failed to generate %s", err)
		return
	}

	ioutil.WriteFile(*outputFile, []byte(out), os.ModePerm)
}
