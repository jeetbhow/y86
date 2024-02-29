package main

import (
	"fmt"
	"log"
	"os"
	"y86/assembler"
)

func main() {
	file, err := os.ReadFile("./file.txt")

	if err != nil {
		log.Fatal(err)
	}

	src := string(file)
	scanner := assembler.NewScanner(src)

	tokens, err := scanner.Scan()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v", tokens)
	}
}
