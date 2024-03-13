package model

import (
	"fmt"
)

// Status codes
const (
	aok byte = iota // All good
	hlt             // Halt instruction encountered
	adr             // Bad address
	ins             // Bad instruction
	dz              // Division by zero
)

// Maps fcodes to ALU functions.
var alu = map[byte]func(int64, int64) int64{
	add: func(valA int64, valB int64) int64 { return valB + valA },
	sub: func(valA int64, valB int64) int64 { return valB - valA },
	and: func(valA int64, valB int64) int64 { return valB & valA },
	xor: func(valA int64, valB int64) int64 { return valB ^ valA },
	mul: func(valA int64, valB int64) int64 { return valB * valA },
	div: func(valA int64, valB int64) int64 { return valB / valA },
	mod: func(valA int64, valB int64) int64 { return valB % valA },
}

const maxMem = 0xffff // Total RAM
const numReg = 16     // Number of registers the CPU supports.
const stackPtrReg = 4 // Stack pointer register

// Output of fetch stage. Important data that's used throughout the pipeline.
type instReg struct {
	opcode byte  // identifier
	fcode  byte  // modifier (OPQ and JXX inst.)
	rA     byte  // source reg
	rB     byte  // dest reg
	valC   int64 // constant
}

// Condition codes. Set after each ALU use.
type cc struct {
	of bool // set after sign overflow
	z  bool // set after zero
	s  bool // set after negative
}

// Container for the CPU state.
type cpuState struct {
	instreg instReg // instruction register
	valP    int     // PC increment if no jump
	valA    int64   // val in rA
	valB    int64   // val in rB
	valE    int64   // ALU output
	valM    int64   // Memory read output
	pc      int     // program counter
	cc      cc      // condition codes
	status  byte    // status register
}

// y86 CPU.
type CPU struct {
	mem   [maxMem]byte  // memory
	reg   [numReg]int64 // registers
	state cpuState      // state
}

func (cpu *CPU) GetMem() *[maxMem]byte {
	return &cpu.mem
}

func (cpu *CPU) GetState() *cpuState {
	return &cpu.state
}

// Advance the clock by one cycle and return the status.
func (cpu *CPU) Tick() byte {
	cpu.fetch()
	cpu.decode()
	cpu.execute()
	cpu.memory()
	cpu.writeback()
	cpu.updatePC()
	return cpu.state.status
}

// Copy a buffer onto a memory location. Return an error if the address is invalid.
func (cpu *CPU) CopyBuf(addr int, buf []byte) error {
	len := len(buf)

	if addr < 0 || addr+len >= int(maxMem) {
		return fmt.Errorf("invalid address %#x", addr)
	}

	for i := 0; i < len; i++ {
		cpu.mem[addr+i] = buf[i]
	}

	return nil
}

// Write a little endiann 8-byte integer to memory at an address. If the address is
// invalid then set the status to ADR.
func (cpu *CPU) writeLongToMem(addr int, val int64) {
	if addr < 0 || addr+8 > maxMem {
		cpu.state.status = adr
		return
	}

	const mask byte = 0xff
	for i := 0; i < 8; i++ {
		byte := (byte(val) & mask)
		cpu.mem[addr+i] = byte
		val = val >> 8
	}
}

func (cpu *CPU) writeBytesToMem(addr int, bytes []byte) error {
	if addr+len(bytes) >= maxMem || addr < 0 {
		return fmt.Errorf("error: cannot write %d bytes to address %#x", len(bytes), addr)
	}

	for i, value := range bytes {
		cpu.mem[addr+i] = value
	}
	return nil
}

// Read a sequence of bytes from memory and store it into a buffer.
func (cpu *CPU) readBytesFromMem(addr int, size int) ([]byte, error) {
	if addr+size >= maxMem || addr < 0 {
		return nil, fmt.Errorf("error: address %#x is invalid", addr)
	}

	bytes := make([]byte, size)
	for i := 0; i < size; i++ {
		bytes[i] = cpu.mem[addr+i]
	}
	return bytes, nil
}

// Return the little endiann 8-byte int at an address. If the address is invalid then set the
// status code to ADR and return 0.
func (cpu *CPU) readMem(addr int) int64 {
	if addr > maxMem-8 {
		cpu.state.status = adr
		return 0
	}

	var val int64 = 0
	for i := 7; i >= 0; i-- {
		byte := cpu.mem[addr+i]
		val = val << 8
		val += int64(byte)
	}

	return val
}

// Write a value to a register.
func (cpu *CPU) writeReg(index byte, val int64) {
	cpu.reg[index] = val
}

// Read a value from a register.
func (cpu *CPU) readReg(index byte) int64 {
	return cpu.reg[index]
}

// Return true if a conditional operation should be carried out and false otherwise.
func (cpu *CPU) ccCheck() bool {
	fcode := cpu.state.instreg.fcode

	switch fcode {
	case 0:
		return true
	case le:
		return cpu.state.cc.z || cpu.state.cc.s
	case l:
		return cpu.state.cc.s
	case e:
		return cpu.state.cc.z
	case ne:
		return !cpu.state.cc.z
	case g:
		return !cpu.state.cc.s
	case ge:
		return !(cpu.state.cc.z || cpu.state.cc.s)
	default:
		cpu.state.status = ins // bad instruction
		return false
	}
}

// Store the hypothetical next value of the program counter in valP.
func (cpu *CPU) setNextPC() {
	opcode := cpu.state.instreg.opcode

	switch opcode {
	case halt:
		cpu.state.valP = cpu.state.pc + 1
	case nop:
		cpu.state.valP = cpu.state.pc + 1
	case rrmovq:
		cpu.state.valP = cpu.state.pc + 2
	case irmovq:
		cpu.state.valP = cpu.state.pc + 10
	case rmmovq:
		cpu.state.valP = cpu.state.pc + 10
	case mrmovq:
		cpu.state.valP = cpu.state.pc + 10
	case opq:
		cpu.state.valP = cpu.state.pc + 2
	case jxx:
		cpu.state.valP = cpu.state.pc + 9
	case call:
		cpu.state.valP = cpu.state.pc + 9
	case ret:
		cpu.state.valP = cpu.state.pc + 1
	case pushq:
		cpu.state.valP = cpu.state.pc + 2
	case popq:
		cpu.state.valP = cpu.state.pc + 2
	default:
		cpu.state.status = ins // bad instruction
		return
	}
}

// Fetch the next instruction and set the instruction register and valP.
func (cpu *CPU) fetch() {
	var size int
	switch cpu.mem[cpu.state.pc] {
	case halt:
		size = 1
	case nop:
		size = 1
	case rrmovq:
		size = 2
	case irmovq:
		size = 10
	case rmmovq:
		size = 10
	case mrmovq:
		size = 10
	case opq:
		size = 2
	case jxx:
		size = 9
	case call:
		size = 9
	case ret:
		size = 1
	case pushq:
		size = 2
	case popq:
		size = 2
	}

	var instruction, err = cpu.readBytesFromMem(cpu.state.pc, size)
	if err != nil {
		cpu.state.status = adr
	}
	cpu.state.instreg = createInstReg(instruction)
	cpu.setNextPC()
}

// Read from the register file and set valA and valB for the execute stage.
func (cpu *CPU) decode() {
	instreg := cpu.state.instreg
	opcode := instreg.opcode

	switch opcode {
	case halt:
		cpu.state.status = hlt
	case nop:
	case irmovq:
		fallthrough
	case rrmovq:
		fallthrough
	case rmmovq:
		fallthrough
	case mrmovq:
		fallthrough
	case jxx:
		fallthrough
	case opq:
		cpu.state.valA = cpu.readReg(instreg.rA)
		cpu.state.valB = cpu.readReg(instreg.rB)
	case call:
		fallthrough
	case ret:
		fallthrough
	case pushq:
		fallthrough
	case popq:
		cpu.state.valA = cpu.readReg(instreg.rA)
		cpu.state.valB = cpu.readReg(stackPtrReg)
	default:
		cpu.state.status = ins
	}
}

// Run the ALU and set valE. Also set the condition codes if an ALU instruction was used.
func (cpu *CPU) execute() {
	opcode := cpu.state.instreg.opcode
	fcode := cpu.state.instreg.fcode

	switch opcode {
	case rrmovq:
		cpu.state.valE = cpu.alu(fcode, cpu.state.valA, 0)
	case irmovq:
		cpu.state.valE = cpu.alu(fcode, cpu.state.instreg.valC, 0)
	case rmmovq:
		cpu.state.valE = cpu.alu(fcode, cpu.state.valB, cpu.state.instreg.valC)
	case mrmovq:
		cpu.state.valE = cpu.alu(fcode, cpu.state.valB, cpu.state.instreg.valC)
	case opq:
		cpu.state.valE = cpu.alu(fcode, cpu.state.valA, cpu.state.valB)
		cpu.updateCC()
	case call:
		cpu.state.valE = cpu.alu(fcode, -8, cpu.state.valB)
	case pushq:
		cpu.state.valE = cpu.alu(fcode, -8, cpu.state.valB)
	case ret:
		cpu.state.valE = cpu.alu(fcode, 8, cpu.state.valB)
	case popq:
		cpu.state.valE = cpu.alu(fcode, 8, cpu.state.valB)
	default:
		return
	}

}

// Returns true if both a and b are negative integers and false otherwise.
func areBothNeg(a int64, b int64) bool {
	return a < 0 && b < 0
}

// Returns true if both a and b are positive integers and false otherwise.
func areBothPos(a int64, b int64) bool {
	return a > 0 && b > 0
}

// Returns true if both a and b are the same sign and false otherwise.
func areSameSign(a int64, b int64) bool {
	return (a < 0) == (b < 0)
}

// Update the condition codes based on the last ALU computation.
func (cpu *CPU) updateCC() {
	if cpu.state.valE == 0 {
		cpu.state.cc.z = true
		return
	}

	if cpu.state.valE < 0 {
		cpu.state.cc.z = true
	}

	switch cpu.state.instreg.fcode {
	case add:
		cpu.state.cc.of = cpu.state.valE > 0 && areBothNeg(cpu.state.valA, cpu.state.valB) || cpu.state.valE < 0 && areBothPos(cpu.state.valA, cpu.state.valB)
	case mul:
		cpu.state.cc.of = cpu.state.valE > 0 && !areSameSign(cpu.state.valA, cpu.state.valB) || cpu.state.valE < 0 && areSameSign(cpu.state.valA, cpu.state.valB)
	case sub:
		cpu.state.cc.of = cpu.state.valE > 0 && cpu.state.valB < 0 && cpu.state.valA > 0 || cpu.state.valE < 0 && cpu.state.valB > 0 && cpu.state.valA < 0
		cpu.state.cc.s = cpu.state.valE < 0
	}
}

// Write or read a value to memory. Set valM if a value was read.
func (cpu *CPU) memory() {
	opcode := cpu.state.instreg.opcode
	valE := int(cpu.state.valE)
	valB := int(cpu.state.valB)
	valP := int64(cpu.state.valP)

	switch opcode {
	case rmmovq:
		cpu.writeLongToMem(valE, cpu.state.valA)
	case mrmovq:
		cpu.state.valM = cpu.readMem(valE)
	case call:
		cpu.writeLongToMem(valE, valP)
	case ret:
		cpu.state.valM = cpu.readMem(valB)
	case pushq:
		cpu.writeLongToMem(valE, cpu.state.valA)
	case popq:
		cpu.state.valM = cpu.readMem(valB)
	default:
		return
	}

}

// Write a value to a register.
func (cpu *CPU) writeback() {
	opcode := cpu.state.instreg.opcode
	instreg := cpu.state.instreg

	switch opcode {
	case rrmovq:
		cpu.writeReg(instreg.rB, cpu.state.valA)
	case mrmovq:
		cpu.writeReg(instreg.rA, cpu.state.valM)
	case call:
		cpu.writeReg(stackPtrReg, cpu.state.valE)
	case ret:
		cpu.writeReg(stackPtrReg, cpu.state.valE)
	case pushq:
		cpu.writeReg(stackPtrReg, cpu.state.valE)
	case popq:
		cpu.writeReg(instreg.rA, cpu.state.valM)
		cpu.writeReg(stackPtrReg, cpu.state.valE)
	case opq:
		cpu.writeReg(instreg.rB, cpu.state.valE)
	default:
		return
	}
}

// Update the value of the PC.
func (cpu *CPU) updatePC() {
	/*
		- Use valM when you get an address from memory (ret)
		- Use valC when you jump to another address (call, jump)
		- Use valP in any other case
	*/
	opcode := cpu.state.instreg.opcode
	valM := int(cpu.state.valM)
	valC := int(cpu.state.instreg.valC)
	valP := int(cpu.state.valP)

	switch opcode {
	case ret:
		cpu.state.pc = valM
	case call:
		cpu.state.pc = valC
	case jxx:
		if cpu.ccCheck() {
			cpu.state.pc = valC
		} else {
			cpu.state.pc = valP
		}
	default:
		cpu.state.pc = valP
	}
}

// Use the ALU to compute valE and return it. Set the Status to Dz if division by 0 is attempted.
func (cpu *CPU) alu(fcode byte, aluA int64, aluB int64) int64 {
	if fcode == div && aluA == 0 {
		cpu.state.status = dz
		return aluB
	}
	op := alu[fcode]
	return op(aluA, aluB)
}
