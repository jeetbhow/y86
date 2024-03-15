package main

import (
	"fmt"
	"os"
	"y86/model"
)

func main() {
	filename := os.Args[1]
	bytes, readError := os.ReadFile(filename)
	source := string(bytes)

	if readError != nil {
		fmt.Println(readError)
	}

	cpu := model.CPU{}
	assembler := *model.NewAssembler(source)
	assemblyError := assembler.Assemble()

	if assemblyError != nil {
		fmt.Println(assemblyError)
	}

	assembler.Load(&cpu)
	cpu.Execute()
	cpu.PrintRegisterFile()
	assembler.PrintDataTable()
}
