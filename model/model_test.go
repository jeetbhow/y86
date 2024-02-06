package model

import "testing"

func TestWriteMem(t *testing.T) {
	cpu := CPU{}
	var addr uint16 = 0x30
	var val int64 = 0x00010203040506070

	success := cpu.WriteMem(addr, val)
	if success != true {
		t.Errorf("write failed when it should have succeed")
	}

	var readVal int64 = 0
	for i := 7; i >= 0; i-- {
		byte := cpu.mem[addr+uint16(i)]
		readVal += int64(byte)
		readVal = readVal << 8
	}
	readVal = readVal >> 8

	if readVal != val {
		t.Errorf("expected %#x but got %#x.", val, readVal)
	}

	for i := 0x38; uint16(i) < maxMem; i++ {
		if cpu.mem[i] != 0 {
			t.Errorf("memory at %#x has a val %#x when it should be 0", uint16(i), cpu.mem[i])
		}
	}

}

func TestWriteMemBadAddr(t *testing.T) {
	cpu := CPU{}
	var addr uint16 = maxMem + 1

	if cpu.WriteMem(addr, 1) != false {
		t.Errorf("wrote to address %#x when size is %#x", addr, maxMem)
	}
}

func TestReadMem(t *testing.T) {
	cpu := CPU{}
	var addr uint16 = 0x30
	bytes := [8]uint8{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	for i := 0; i < 8; i++ {
		cpu.mem[addr+uint16(i)] = bytes[i]
	}

	val, err := cpu.ReadMem(addr)
	var expectedVal int64 = 0x0706050403020100

	if err != nil {
		t.Error(err)
	}

	if val != expectedVal {
		t.Errorf("expected %#x but got %#x", expectedVal, val)
	}
}

func TestReadMemBadAddr(t *testing.T) {
	cpu := CPU{}
	var addr uint16 = 1017

	_, err := cpu.ReadMem(addr)

	if err == nil {
		t.Errorf("read succeeded when it should have failed")
	}
}
