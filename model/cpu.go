package model

import (
	"fmt"
)

/*
	ALU instructions
	0 - addq
	1 - subq
	2 - andq
	3 - xorq
	4 - mulq
	5 - divq
	6 - modq
*/

var ALU = map[uint8]func(int64, int64) int64{
	0: func(valA int64, valB int64) int64 { return valB + valA },
	1: func(valA int64, valB int64) int64 { return valB - valA },
	2: func(valA int64, valB int64) int64 { return valB & valA },
	3: func(valA int64, valB int64) int64 { return valB ^ valA },
	4: func(valA int64, valB int64) int64 { return valB * valA },
	5: func(valA int64, valB int64) int64 { return valB / valA },
	6: func(valA int64, valB int64) int64 { return valB % valA },
}

// Max working memory.
const maxMem uint16 = 1024

// Output of fetch stage. Important instruction data that's used throughout the pipeline.
type InstReg struct {
	icode uint8 // identifier
	ifun  uint8 // modifier (OPQ and JXX inst.)
	rA    uint8 // source reg
	rB    uint8 // dest reg
	valC  int64 // constant
	valP  int64 // PC increment if no jump
}

// Condition codes. Set after each ALU tick.
type CC struct {
	of bool // set after sign overflow
	z  bool // set after zero
	s  bool // set after negative
}

// y86 CPU.
type CPU struct {
	mem    [maxMem]uint8 // memory
	reg    [15]int64     // general purpose registers
	pc     uint32        // program counter
	cc     CC            // condition codes
	status uint8         // status register
}

// Write a little-endiann 8-byte integer to memory at an address. Returns true if the write
// was successful, and false otherwise.
func (cpu *CPU) WriteMem(addr uint16, val int64) bool {
	if addr+8 > maxMem {
		return false
	}

	const mask = 0xff
	for i := 0; i < 8; i++ {
		byte := (uint8(val) & mask)
		cpu.mem[addr+uint16(i)] = byte
		val = val >> 8
	}

	return true
}

// Return the 8-byte int at an address. Return an error if the address is invalid.
// Check the error before using the value when calling this function as the value
// will not be correct.
func (cpu *CPU) ReadMem(addr uint16) (int64, error) {
	if addr > maxMem-8 {
		return 0, fmt.Errorf("attempting to read from %#x to %#x, but mem size is %#x", addr, addr+7, maxMem)
	}

	var val int64 = 0
	for i := 7; i >= 0; i-- {
		byte := cpu.mem[addr+uint16(i)]
		val = val << 8
		val += int64(byte)
	}

	return val, nil
}

// TODO: Write a value to a register.
func (cpu *CPU) WriteReg(reg uint8, val int64) {

}

// TODO: Read a value from a register.
func ReadReg(reg uint8) int64 {
	return 0
}

// TODO: Fetch the next instruction using the PC. Returns the InstReg object
// corresponding to the instruction.
func (cpu *CPU) Fetch() (*InstReg, error) {
	return nil, nil
}

// TODO: Get valA and valB from the instruction.
func (cpu *CPU) Decode(instreg *InstReg) (uint8, uint8) {
	return 0, 0
}

// TODO: Use the ALU and get valE. Also set the condition codes.
func (cpu *CPU) Execute(ifun uint8, valA int64, valB int64) int64 {
	return 0
}

// TODO: Write or read a value to memory.
func (cpu *CPU) Memory() {

}

// TODO Write a value to a register.
func (cpu *CPU) Writeback() {

}
