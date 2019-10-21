package vm

import (
	"golox/repr"
	"time"
)

func (vm *VM) defineNative(name string, nativeFn repr.NativeFn) {
	vm.Globals[name] = repr.NativeVal(nativeFn)
}

func (vm *VM) initNatives() {
	vm.defineNative("clock", clockNative)
}

func clockNative(argCount int, args []repr.Value) repr.Value {
	return repr.NumberVal(float64(time.Now().Unix()))
}
