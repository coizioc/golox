package vm

import (
	"fmt"
	"golox/chunk"
	"golox/parser"
)

type InterpretResult byte

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Chunk *chunk.Chunk
	IP    byte
	Stack []chunk.Value
}

func New() *VM {
	return &VM{nil, 0, []chunk.Value{}}
}

func (vm *VM) Interpret(source string) InterpretResult {
	c := chunk.New()

	if !parser.Compile(source, c) {
		return INTERPRET_COMPILE_ERROR
	}

	vm.Chunk = c
	vm.IP = vm.Chunk.Code[0]
	return vm.run()
}

func (vm *VM) push(value chunk.Value) {
	vm.Stack = append(vm.Stack, value)
}

func (vm *VM) pop() chunk.Value {
	value := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return value
}

func (vm *VM) binaryOp(op byte) {
	b, a := vm.pop(), vm.pop()

	switch op {
	case chunk.OP_ADD:
		vm.push(a + b)
	case chunk.OP_SUBTRACT:
		vm.push(a - b)
	case chunk.OP_MULTIPLY:
		vm.push(a * b)
	case chunk.OP_DIVIDE:
		vm.push(a / b)
	}
}

func (vm *VM) readByte() byte {
	byteRead := vm.Chunk.Code[vm.IP]
	vm.IP++
	return byteRead
}

func (vm *VM) readConstant() chunk.Value {
	byteRead := vm.readByte()
	return vm.Chunk.Constants[byteRead]
}

func (vm *VM) run() InterpretResult {
	for {
		instruction := vm.readByte()
		switch instruction {
		case chunk.OP_CONSTANT:
			constant := vm.readConstant()
			vm.push(constant)
		case chunk.OP_ADD, chunk.OP_SUBTRACT, chunk.OP_MULTIPLY, chunk.OP_DIVIDE:
			vm.binaryOp(instruction)
		case chunk.OP_NEGATE:
			vm.push(-vm.pop())
		case chunk.OP_RETURN:
			fmt.Println(vm.pop())
			return INTERPRET_OK
		}
	}
}
