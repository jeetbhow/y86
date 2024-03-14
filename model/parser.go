package model

import (
	"fmt"
	"strconv"
)

// Contains functions for parsing an instruction and converting it into a byte representation.
var parseDispatchTable = map[byte]func(token Token, opcode byte, fcode byte, size byte, parser *Parser) error{
	halt:   parse1Byte,
	nop:    parse1Byte,
	opq:    parse2Byte,
	irmovq: parseIrmovq,
	mrmovq: parseMrmovq,
}

// Object that converts a list of tokens to a set of machine instructions which it can save on the disk.
type Parser struct {
	tokens            []Token        // list of tokens
	curr              int            // the current token index
	symbolTable       map[string]int // contains all of the labels and their addresses
	dataTable         map[int]int64  // contains all of the data to be stored in memory
	instructionBuffer [][]byte       // contains machine code
	start             int            // the starting address of the program
	lc                int            // location counter
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:            tokens,
		symbolTable:       make(map[string]int),
		dataTable:         make(map[int]int64),
		instructionBuffer: make([][]byte, 0),
	}
}

func (p *Parser) GetDataTable() map[int]int64 {
	return p.dataTable
}

func (p *Parser) GetInstructionBuffer() [][]byte {
	return p.instructionBuffer
}

// Print the symbol table to the console.
func (p *Parser) PrintSymbolTable() {
	fmt.Println(p.symbolTable)
}

// Print the data table to the console.
func (p *Parser) PrintDataTable() {
	fmt.Println(p.dataTable)
}

// Print the machine code to the console.
func (p *Parser) PrintInsBuf() {
	fmt.Printf("%x\n", p.instructionBuffer)
}

// Return the starting address of the program.
func (p *Parser) GetStart() int {
	return p.start
}

// Create the machine code translation of the assembly code.
func (p *Parser) Parse() error {
	err1 := p.firstPass()
	err2 := p.secondPass()

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

	return nil
}

// The first pass through the token list will construct the symbol and data tables. The reason
// a first pass is necessary is because in code where the instructions are laid out before
// the label declarations, there's no way to figure out what address of those labels.
func (p *Parser) firstPass() error {
	for !p.isAtEnd() {
		currToken := p.advance()

		switch currToken.tokenType {
		case dir:
			err := p.parseDirective(currToken)
			if err != nil {
				return err
			}
		case instruction:
			p.lc += int(instructionTable[currToken.lex][2])
		case label:
			if next := p.peek(); next.tokenType == colon {
				p.symbolTable[currToken.lex] = p.lc
			}
		}
	}
	p.curr = 0
	return nil
}

// The second pass through the token list will generate the obj file containing the
// machine code for the instructions.
func (p *Parser) secondPass() error {
	for !p.isAtEnd() {
		currToken := p.advance()

		switch currToken.tokenType {
		case dir:
			p.parseDirective(currToken)
		case instruction:
			p.parseInstruction(currToken)
		}
	}
	return nil
}

func (l *Parser) LoadCPU(cpu *CPU) {
	l.setEntryPoint(cpu)
	l.loadData(cpu)
	l.loadInstructions(cpu)
}

func (l *Parser) setEntryPoint(cpu *CPU) {
	cpu.state.pc = l.start
}

func (l *Parser) loadData(cpu *CPU) {
	for address, data := range l.dataTable {
		cpu.writeLongToMem(address, data)
	}
}

func (l *Parser) loadInstructions(cpu *CPU) error {
	address := cpu.state.pc
	for _, bytes := range l.instructionBuffer {
		err := cpu.writeBytesToMem(address, bytes)
		if err != nil {
			return err
		}
		address += len(bytes)
	}
	return nil
}

// Returns true if there are no more tokens left to parse and false if there are.
func (p *Parser) isAtEnd() bool {
	return p.tokens[p.curr].tokenType == eof
}

// Return the current token and then advance the parser.
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.curr++
	}
	return p.tokens[p.curr-1]
}

// Return the current token without advancing the parser.
func (p *Parser) peek() Token {
	return p.tokens[p.curr]
}

// Assuming that the token is a directive, this function will figure out what
// kind of directive it is and what the assembler should do in response.
func (p *Parser) parseDirective(token Token) error {
	/*
		The two directives in the y86 assembly language are .pos and .quad.
		Both of these directives require a number as the next token.
		The .pos directive updates the location counter whereas the .quad
		directive tells the assembler to store something in memory.
	*/
	next := p.advance()
	if next.tokenType != num {
		return fmt.Errorf("invalid directive at [%d:%d]: expected number, got %s", next.line, next.col, next.lex)
	}

	switch token.lex {
	case ".pos":
		address, _ := strconv.ParseInt(next.lex, 0, 0)
		// this sets the starting address of the program if it hasn't been set yet.
		if p.start == 0 && p.peek().tokenType == instruction {
			p.start = int(address)
		}
		p.lc = int(address)
	case ".quad":
		val, _ := strconv.ParseInt(next.lex, 0, 0)
		p.dataTable[p.lc] = val
		p.lc += 8
	}
	return nil
}

// Assuming that the token is an instruction, this function will figure out what
// kind of instruction it is and what the assembler should do in response.
func (p *Parser) parseInstruction(token Token) error {
	var instructionInfo []byte = instructionTable[token.lex]
	opcode := instructionInfo[0]
	fcode := instructionInfo[1]
	size := instructionInfo[2]
	err := parseDispatchTable[opcode](token, opcode, fcode, size, p)
	if err != nil {
		return err
	}
	return nil
}

// Parses a 1 byte instruction such as halt, nop, ret
var parse1Byte = func(token Token, opcode byte, fcode byte, size byte, p *Parser) error {
	instruction := make([]byte, size)
	instruction[0] = opcode<<4 | fcode
	p.instructionBuffer = append(p.instructionBuffer, instruction)
	return nil
}

// Parses a 2 bytes instruction such as opq, pushq, popq. They have the form Command-Reg-Reg
var parse2Byte = func(token Token, opcode byte, fcode byte, size byte, p *Parser) error {
	instruction := make([]byte, size)
	args := []Token{p.advance(), p.advance(), p.advance()}

	if IsEof(args) {
		return fmt.Errorf("unexpected eof at [%d:%d]", token.line, token.col)
	} else if !IsValidArgs(args, reg, comma, reg) {
		return fmt.Errorf("invalid arguments at [%d:%d]", token.line, token.col)
	}
	rA, rAExists := registerTable[args[0].lex]
	rB, rBExists := registerTable[args[2].lex]

	if !rAExists {
		return fmt.Errorf("invalid register at [%d:%d]", args[0].line, args[0].col)
	} else if !rBExists {
		return fmt.Errorf("invalid register at [%d:%d]", args[2].line, args[2].col)
	}

	instruction[1] = rA<<4 | rB
	p.instructionBuffer = append(p.instructionBuffer, instruction)
	return nil
}

// Parse the irmovq instruction and increment the location counter of the parser.
var parseIrmovq = func(token Token, opcode byte, fcode byte, size byte, p *Parser) error {
	var args = []Token{p.advance(), p.advance(), p.advance()}
	bytes := make([]byte, size)

	if IsEof(args) {
		return fmt.Errorf("unexpected eof at [%d:%d]", token.line, token.col)
	} else if !IsValidArgs(args, label, comma, reg) && !IsValidArgs(args, num, comma, reg) {
		return fmt.Errorf("invalid arguments at [%d:%d]", token.line, token.col)
	}
	bytes[0] = byte(opcode<<4 | fcode)
	var rA byte = 0xf
	rB, ok := registerTable[args[2].lex]
	if !ok {
		return fmt.Errorf("invalid register at [%d:%d]", args[2].line, args[2].col)
	}

	bytes[1] = byte(rA<<4 | rB)

	switch args[0].tokenType {
	case num:
		val, _ := strconv.ParseInt(args[0].lex, 0, 0)
		copy(bytes[2:], intToBytes(val))
	case label:
		val := p.symbolTable[args[0].lex]
		copy(bytes[2:], intToBytes(int64(val)))
	}
	p.instructionBuffer = append(p.instructionBuffer, bytes)
	p.lc += int(size)
	return nil
}

var parseMrmovq = func(token Token, opcode byte, fcode byte, size byte, p *Parser) error {
	var args = make([]Token, 5)
	var valC int64
	var rA byte
	var rB byte
	var bytes = make([]byte, size)

	bytes[0] = byte(opcode<<4 | fcode)

	args[0] = p.advance()
	switch args[0].tokenType {
	case lparen:
		valC = 0
	case num:
		valC, _ = strconv.ParseInt(args[0].lex, 0, 0)
		args[0] = p.advance() // We advance the parser to account for the offset present in the instruction.
	case eof:
		return fmt.Errorf("unexpected eof at [%d:%d]", token.line, token.col)
	default:
		return fmt.Errorf("invalid arguments at [%d:%d]", token.line, token.col)
	}

	for i := 1; i < len(args); i++ {
		args[i] = p.advance()
	}

	if IsEof(args) {
		return fmt.Errorf("unexpected eof at [%d:%d]", token.line, token.col)
	}

	if !IsValidArgs(args, lparen, reg, rparen, comma, reg) {
		return fmt.Errorf("invalid arguments at [%d:%d]", token.line, token.col)
	}

	rB = registerTable[args[1].lex]
	rA = registerTable[args[4].lex]
	bytes[1] = byte(rA<<4 | rB)
	copy(bytes[2:], intToBytes(valC))
	p.instructionBuffer = append(p.instructionBuffer, bytes)
	p.lc += int(size)
	return nil
}
