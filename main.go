package main

import "y86/model"

func main() {
	cpu := model.CPU{}

	// instructions
	cpu.CopyBuf(0, model.EncodeInst(1, 0, 0, 0, 0))
	cpu.CopyBuf(1, model.EncodeInst(0, 0, 0, 0, 0))

	var status byte = 0
	for status == 0 {
		status = cpu.Tick()
	}
}
