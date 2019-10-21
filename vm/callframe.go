package vm

import "golox/repr"

type CallFrame struct {
	Function   *repr.Function
	IP         int
	StackStart int
}

func (vm *VM) AddFrame(function *repr.Function, ip, stackStart int) {
	vm.Frames = append(vm.Frames, &CallFrame{function, ip, stackStart})
}

func (vm *VM) RemoveFrame() {
	vm.Stack = vm.Stack[:vm.CurrFrame().StackStart]
	vm.Frames = vm.Frames[:vm.FrameCount()-1]
}
