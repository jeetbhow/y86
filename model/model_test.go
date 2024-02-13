package model

import (
	"bytes"
	"math"
	"testing"
)

var haltState cpuState = cpuState{
	valP:   1,
	pc:     1,
	status: hlt,
}

var nopState cpuState = cpuState{
	instreg: instReg{
		opcode: 1,
	},
	valP: 1,
	pc:   1,
}

var rrmovqState cpuState = cpuState{
	instreg: instReg{
		opcode: 2,
		rA:     1,
		rB:     2,
	},
	valA: 1,
	valB: 2,
	valE: 1,
	valP: 2,
	pc:   2,
}

var irmovqState cpuState = cpuState{
	instreg: instReg{
		opcode: 3,
		rA:     15,
		rB:     2,
		valC:   0xffff,
	},
	valA: 15,
	valB: 2,
	valE: 0xffff,
	valP: 10,
	pc:   10,
}

var rmmovqState cpuState = cpuState{
	instreg: instReg{
		opcode: 4,
		rA:     1,
		rB:     2,
		valC:   2,
	},
	valA: 1,
	valB: 2,
	valE: 4,
	valP: 10,
	pc:   10,
}

var mrmovqState cpuState = cpuState{
	instreg: instReg{
		opcode: 5,
		rA:     1,
		rB:     2,
		valC:   2,
	},
	valA: 1,
	valB: 2,
	valE: 4,
	valP: 10,
	pc:   10,
}

var aluAddState cpuState = cpuState{
	instreg: instReg{
		opcode: 6,
		fcode:  0,
		rA:     1,
		rB:     2,
	},
	valA: 1,
	valB: 2,
	valE: 3,
	valP: 2,
	pc:   2,
}

var jumpState cpuState = cpuState{
	instreg: instReg{
		opcode: 7,
		fcode:  ne,
		valC:   0x2000,
	},
	valP: 9,
	pc:   0x2000,
}

var callState cpuState = cpuState{
	instreg: instReg{
		opcode: 8,
		valC:   0x2000,
	},
	valB: 0x30,
	valE: 0x28,
	valP: 9,
	pc:   0x2000,
}

var retState cpuState = cpuState{
	instreg: instReg{
		opcode: 9,
	},
	valB: 0x30,
	valE: 0x38,
	valP: 1,
	valM: 0x2000,
	pc:   0x2000,
}

var pushState cpuState = cpuState{
	instreg: instReg{
		opcode: 10,
		rA:     0x1,
		rB:     0xf,
	},
	valA: 0x1,
	valB: 0x30,
	valE: 0x28,
	valP: 2,
	pc:   0x2,
}

var popState cpuState = cpuState{
	instreg: instReg{
		opcode: 11,
		rA:     0x1,
		rB:     0xf,
	},
	valA: 0x1,
	valB: 0x30,
	valE: 0x38,
	valP: 2,
	valM: 0x2000,
	pc:   0x2,
}

func TestWriteMem(t *testing.T) {
	cpu := CPU{}
	addr := 0
	var val int64 = 0x00010203040506070
	cpu.writeMem(addr, val)

	var readVal int64 = 0
	for i := 7; i >= 0; i-- {
		byte := cpu.mem[addr+i]
		readVal += int64(byte)
		readVal = readVal << 8
	}
	readVal = readVal >> 8

	if readVal != val {
		t.Errorf("expected %#x but got %#x\n", val, readVal)
	}

	for i := 8; uint16(i) < maxMem; i++ {
		if cpu.mem[i] != 0 {
			t.Errorf("memory at %#x has a val %#x when it should be 0\n", uint16(i), cpu.mem[i])
		}
	}

}

func TestWriteMemBadAddr(t *testing.T) {
	cpu := CPU{}
	addr := maxMem + 1
	cpu.writeMem(addr, 0)

	if cpu.state.status != adr {
		t.Fail()
	}
}

func TestReadMem(t *testing.T) {
	cpu := CPU{}
	addr := 0
	buf := []uint8{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	cpu.CopyBuf(addr, buf)

	val := cpu.readMem(addr)
	var expected int64 = 0x0706050403020100

	if val != expected {
		t.Errorf("expected %#x but got %#x\n", expected, val)
	}
}

func TestReadMemBadAddr(t *testing.T) {
	cpu := CPU{}
	addr := maxMem - 3
	cpu.readMem(addr)

	if cpu.state.status != adr {
		t.Error("expected Adr status")
	}
}

func TestToBytes(t *testing.T) {
	var val int64 = 0x0102030405060708
	var res []byte = intToBytes(val)
	expected := []byte{0x8, 0x7, 0x6, 0x5, 0x4, 0x3, 0x2, 0x1}

	if !bytes.Equal(res, expected) {
		t.Errorf("expected:  %v actual: %v\n", expected, res)
	}
}

func TestToInt(t *testing.T) {
	buf := []byte{0x8, 0x7, 0x6, 0x5, 0x4, 0x3, 0x2, 0x1}
	var expected int64 = 0x0102030405060708
	res := bytesToInt(buf)

	if res != expected {
		t.Errorf("expected %#x but got %#x\n", expected, res)
	}
}

func TestCCOverflow(t *testing.T) {
	testcases := []struct {
		name     string
		valE     int64
		valA     int64
		valB     int64
		fcode    byte
		expected bool
	}{
		{"add pos overflow", math.MinInt64, math.MaxInt64, 1, add, true},
		{"add neg overflow", math.MaxInt64, math.MinInt64, -1, add, true},
		{"add no overflow pos", 2, 1, 1, add, false},
		{"add no overflow neg", -2, -1, -1, add, false},
		{"sub pos overflow", math.MaxInt64, 1, math.MinInt64, sub, true},
		{"sub neg overflow", math.MinInt64, -1, math.MaxInt64, sub, true},
		{"sub no overflow pos", 1, 1, 2, sub, false},
		{"sub no overflow neg", -3, 1, -2, sub, false},
		{"mul same sign overflow1", -3037000499, 3037000500, 3037000500, mul, true},
		{"mul same sign overflow2", -3037000499, -3037000500, -3037000500, mul, true},
		{"mul diff sign overflow1", 3037000499, 3037000500, -3037000500, mul, true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cpu := CPU{}
			cpu.state.valE = tc.valE
			cpu.state.valA = tc.valA
			cpu.state.valB = tc.valB
			cpu.state.instreg.fcode = tc.fcode

			cpu.updateCC()
			if cpu.state.cc.of != tc.expected {
				t.Fail()
			}
		})
	}
}

func TestTick(t *testing.T) {
	testcases := []struct {
		name          string
		inst          []byte
		expectedState cpuState
	}{
		{"halt", EncodeInst(0, 0, 0, 0, 0), haltState},
		{"nop", EncodeInst(1, 0, 0, 0, 0), nopState},
		{"rrmovq", EncodeInst(2, 0, 1, 2, 0), rrmovqState},
		{"irmovq", EncodeInst(3, 0, 0xf, 2, 0xffff), irmovqState},
		{"rmmovq", EncodeInst(4, 0, 1, 2, 2), rmmovqState},
		{"mrmovq", EncodeInst(5, 0, 1, 2, 2), mrmovqState},
		{"add", EncodeInst(6, 0, 1, 2, 0), aluAddState},
		{"jump", EncodeInst(7, ne, 0, 0, 0x2000), jumpState},
		{"call", EncodeInst(8, 0, 0, 0, 0x2000), callState},
		{"ret", EncodeInst(9, 0, 0, 0, 0), retState},
		{"push", EncodeInst(10, 0, 0x1, 0xf, 0), pushState},
		{"pop", EncodeInst(11, 0, 0x1, 0xf, 0), popState},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cpu := CPU{}

			for i := 0; i < 16; i++ {
				cpu.reg[i] = int64(i)
			}

			addr := 0
			cpu.CopyBuf(addr, tc.inst)
			cpu.writeReg(stackPtrReg, 0x30)
			cpu.writeMem(0x30, 0x2000)
			cpu.Tick()
			if cpu.state != tc.expectedState {
				t.Errorf("expected %+v but got %+v\n", tc.expectedState, cpu.state)
			}
		})
	}
}
