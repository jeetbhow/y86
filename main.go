package main

import (
	"fmt"
	"log"
	"os"
	"y86/model"
)

func main() {
	file, err := os.ReadFile("./file.txt")

	if err != nil {
		log.Fatal(err)
	}

	src := string(file)
	scanner := model.NewScanner(src)

	tokens, err := scanner.Scan()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", tokens)
	}

	parser := model.NewParser(tokens)
	err = parser.Parse()

	if err != nil {
		panic(err)
	}

	parser.PrintSymbolTable()
	parser.PrintDataTable()
	parser.PrintInsBuf()
}
