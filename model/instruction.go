package model

const (
	halt   byte = iota // Halt cpu
	nop                // Do nothing
	rrmovq             // Move values between registers
	irmovq             // Move a constant immediately to a register
	rmmovq             // Move data from register to memory
	mrmovq             // Move data from memory to register
	opq                // Alu (op could be add, sub, mul, div, and, xor, mod)
	jxx                // Jump (xx could be le, l, e, ne, g, ge)
	call               // Function call
	ret                // Return to caller
	pushq              // Push onto call stack
	popq               // Pop from call stack
)

// Alu flags
const (
	add byte = iota
	sub
	and
	xor
	mul
	div
	mod
)

// Conditional check flags
const (
	le = sub
	l  = and
	e  = xor
	ne = mul
	ge = div
	g  = mod
)

// Convert an 8 byte integer to a byte slice in little endiann format.
func intToBytes(val int64) []byte {
	const mask byte = 0xff
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		byte := byte(val) & mask
		res[i] = byte
		val = val >> 8
	}
	return res
}

// Extract the 8 byte little-endiann integer from a byte slice.
func bytesToInt(bytes []byte) int64 {
	var res int64 = 0
	for i := len(bytes) - 1; i >= 0; i-- {
		res = res << 8
		res = res | int64(bytes[i])
	}
	return res
}

// Create an instruction register from a byte slice.
func createInstReg(bytes []byte) instReg {
	opcode := (bytes[0] & 0xf0) >> 4
	fcode := bytes[0] & 0x0f

	switch opcode {
	case halt:
		fallthrough
	case nop:
		fallthrough
	case ret:
		return instReg{
			opcode: opcode,
			fcode:  fcode,
		}
	case jxx:
		fallthrough
	case call:
		return instReg{
			opcode: opcode,
			fcode:  fcode,
			valC:   bytesToInt(bytes[1:8]),
		}
	case irmovq:
		fallthrough
	case rmmovq:
		fallthrough
	case mrmovq:
		return instReg{
			opcode: opcode,
			fcode:  fcode,
			rA:     (bytes[1] & 0xf0) >> 4,
			rB:     bytes[1] & 0x0f,
			valC:   bytesToInt(bytes[2:8]),
		}
	default:
		return instReg{
			opcode: opcode,
			fcode:  fcode,
			rA:     (bytes[1] & 0xf0) >> 4,
			rB:     bytes[1] & 0x0f,
		}
	}
}

// Construct a byte slice representation of the instruction using the instruction register paramters. Panics
// if the parameters do not encode a valid instruction.
func EncodeInst(opcode byte, fcode byte, rA byte, rB byte, constant int64) []byte {
	if fcode > 6 {
		panic("invalid instruction")
	}

	iByte := opcode<<4 | fcode
	regByte := rA<<4 | rB

	justIByte := []byte{iByte}
	iByteAndReg := []byte{iByte, regByte}
	iByteAndConst := append([]byte{iByte}, intToBytes(constant)...)
	everything := append([]byte{iByte, regByte}, intToBytes(constant)...)

	switch opcode {
	case halt:
		return justIByte
	case nop:
		return justIByte
	case rrmovq:
		return iByteAndReg
	case irmovq:
		return everything
	case rmmovq:
		return everything
	case mrmovq:
		return everything
	case opq:
		return iByteAndReg
	case jxx:
		return iByteAndConst
	case call:
		return iByteAndConst
	case ret:
		return justIByte
	case pushq:
		return iByteAndReg
	case popq:
		return iByteAndReg
	default:
		panic("invalid instruction")
	}
}
