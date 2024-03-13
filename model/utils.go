package model

/*
 * This file contains utility functions and static data used throughout
 * the project.
 */

// Table of lexemes and their respective token types.
var lexemeTable = map[string]TokenType{
	"halt":   instruction,
	"nop":    instruction,
	"rrmovq": instruction,
	"irmovq": instruction,
	"rmmovq": instruction,
	"mrmovq": instruction,
	"addq":   instruction,
	"subq":   instruction,
	"andq":   instruction,
	"xorq":   instruction,
	"mulq":   instruction,
	"divq":   instruction,
	"modq":   instruction,
	"jmp":    instruction,
	"jle":    instruction,
	"jl":     instruction,
	"je":     instruction,
	"jne":    instruction,
	"jge":    instruction,
	"jg":     instruction,
	"call":   instruction,
	"ret":    instruction,
	"pushq":  instruction,
	"popq":   instruction,
	".pos":   dir,
	".quad":  dir,
}

// Table of register strings and their numberical values.
var registerTable = map[string]byte{
	"%rax": 0,
	"%rcx": 1,
	"%rdx": 2,
	"%rbx": 3,
	"%rsp": 4,
	"%rbp": 5,
	"%rsi": 6,
	"%rdi": 7,
	"%r8":  8,
	"%r9":  9,
	"%r10": 10,
	"%r11": 11,
	"%r12": 12,
	"%r13": 13,
	"%r14": 14,
	"%r15": 15,
}

// Maps instruction strings to their unique identifiers. This includes the opcode, fcode, and size.
var instructionTable = map[string][]byte{
	"halt":   {0, 0, 1},
	"nop":    {1, 0, 1},
	"rrmovq": {2, 0, 2},
	"irmovq": {3, 0, 10},
	"rmmovq": {4, 0, 10},
	"mrmovq": {5, 0, 10},
	"addq":   {6, 0, 2},
	"subq":   {6, 1, 2},
	"andq":   {6, 2, 2},
	"xorq":   {6, 3, 2},
	"mulq":   {6, 4, 2},
	"divq":   {6, 4, 2},
	"modq":   {6, 5, 2},
	"jmp":    {7, 0, 9},
	"jle":    {7, 1, 9},
	"jl":     {7, 2, 9},
	"je":     {7, 3, 9},
	"jne":    {7, 4, 9},
	"jge":    {7, 5, 9},
	"jg":     {7, 6, 9},
	"call":   {8, 0, 9},
	"ret":    {9, 0, 1},
	"pushq":  {10, 0, 2},
	"popq":   {11, 0, 2},
}

// Returns true if all the tokens are eof and false otherwise.
func IsEof(tokens []Token) bool {
	for _, t := range tokens {
		if t.tokenType != eof {
			return false
		}
	}
	return true
}

// Takes a list of tokens and matches them against the expected token types. Returns true if the tokens match
// the expected types and false otherwise.
func IsValidArgs(args []Token, expected ...TokenType) bool {
	// not enough arguments
	if (len(args) != len(expected)) || (len(args) == 0) {
		return false
	}

	for i := 0; i < len(args); i++ {
		if args[i].tokenType != expected[i] {
			return false
		}
	}
	return true
}
