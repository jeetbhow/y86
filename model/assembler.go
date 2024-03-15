package model

import "fmt"

type Assembler struct {
	scanner Scanner
	parser  Parser
}

// Create a new assembler and set the source string to assemble.
func NewAssembler(src string) *Assembler {
	scanner := NewScanner(src)
	return &Assembler{
		*scanner,
		Parser{symbolTable: make(map[string]int), dataTable: make(map[int]int64)},
	}
}

// Assemble the source code and generate the instruction buffer. Return an error if
// an error occurred in either the scanning or parsing phase.
func (a *Assembler) Assemble() error {
	scanError := a.scanner.scan()
	a.parser.SetTokens(a.scanner.tokens)
	parseError := a.parser.parse()
	if scanError != nil {
		return scanError
	} else if parseError != nil {
		return parseError
	} else {
		return nil
	}
}

// Print the instrution buffer.
func (a *Assembler) PrintInstructions() {
	fmt.Println(a.parser.instructions)
}

// Print the data table
func (a *Assembler) PrintDataTable() {
	fmt.Println("Data Memory:")
	for address, data := range a.parser.dataTable {
		fmt.Printf("%#x: %d\n", address, data)
	}
}

// Load the data table and instruction buffer into the CPU.
func (a *Assembler) Load(cpu *CPU) error {
	a.setEntryPoint(cpu)
	dataError := a.loadData(cpu)
	instructionError := a.loadInstructions(cpu)
	if dataError != nil {
		return dataError
	} else if instructionError != nil {
		return instructionError
	} else {
		return nil
	}
}

// Set the program counter to the entry point of the CPU.
func (a *Assembler) setEntryPoint(cpu *CPU) {
	cpu.state.pc = a.parser.start
}

// Load the instruction buffer into the CPU starting at the location of the program counter.
func (a *Assembler) loadInstructions(cpu *CPU) error {
	address := cpu.state.pc
	for _, bytes := range a.parser.instructions {
		err := cpu.writeBytesToMem(address, bytes)
		if err != nil {
			return err
		}
		address += len(bytes)
	}
	return nil
}

// Load the data into memory.
func (a *Assembler) loadData(cpu *CPU) error {
	var dataTable map[int]int64 = a.parser.dataTable
	for address, value := range dataTable {
		err := cpu.writeLongToMem(address, value)
		if err != nil {
			return err
		}
	}
	return nil
}
