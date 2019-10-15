package vm

import (
	"fmt"
	"golox/chunk"
	"golox/loxerror"
	"golox/parser"
	"golox/value"
)

type InterpretResult byte

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Chunk   *chunk.Chunk
	IP      int
	Stack   []value.Value
	Globals map[string]value.Value
	Out     interface{}
}

func New() *VM {
	return &VM{nil, 0, []value.Value{}, make(map[string]value.Value), nil}
}

func (vm *VM) Interpret(source string) InterpretResult {
	c := chunk.New()
	p := parser.New(source, c)

	if !p.Compile() {
		return INTERPRET_COMPILE_ERROR
	}

	vm.Chunk = c
	vm.IP = 0
	return vm.run()
}

func (vm *VM) push(value value.Value) {
	vm.Stack = append(vm.Stack, value)
}

func (vm *VM) pop() value.Value {
	retValue := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return retValue
}

func (vm *VM) peek(distance int) value.Value {
	return vm.Stack[len(vm.Stack)-1-distance]
}

/*
func (vm *VM) line() int {
	instruction := vm.Chunk.Code[vm.IP]
	line := vm.Chunk.Code[instruction]
}
*/

func (vm *VM) binaryOp(op byte) {
	var a, b float64
	byteb, bytea := vm.peek(0), vm.peek(1)
	if op == chunk.OP_ADD && bytea.IsString() && byteb.IsString() {
		vm.concatenate()
		return
	} else if bytea.IsNumber() && byteb.IsNumber() {
		b, a = vm.pop().AsNumber(), vm.pop().AsNumber()
	} else {
		loxerror.Error(-1, "Operands must be a number")
	}

	switch op {
	case chunk.OP_GREATER:
		vm.push(value.BoolVal(a > b))
	case chunk.OP_LESS:
		vm.push(value.BoolVal(a < b))
	case chunk.OP_ADD:
		vm.push(value.NumberVal(a + b))
	case chunk.OP_SUBTRACT:
		vm.push(value.NumberVal(a - b))
	case chunk.OP_MULTIPLY:
		vm.push(value.NumberVal(a * b))
	case chunk.OP_DIVIDE:
		vm.push(value.NumberVal(a / b))
	}
}

func (vm *VM) concatenate() {
	b, a := vm.pop().AsString(), vm.pop().AsString()

	vm.push(value.StringVal(a + b))
}

func (vm *VM) isFalsey(v value.Value) bool {
	return v.IsNil() || (v.IsBool() && !v.AsBool())
}

func (vm *VM) readByte() byte {
	byteRead := vm.Chunk.Code[vm.IP]
	vm.IP++
	return byteRead
}

func (vm *VM) readConstant() value.Value {
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
		case chunk.OP_NIL:
			vm.push(value.NilVal())
		case chunk.OP_TRUE:
			vm.push(value.BoolVal(true))
		case chunk.OP_FALSE:
			vm.push(value.BoolVal(false))
		case chunk.OP_POP:
			vm.pop()
		case chunk.OP_GET_GLOBAL:
			name := vm.readConstant().AsString()
			val, ok := vm.Globals[name]
			if !ok {
				loxerror.Error(-1, "Undefined variable '"+name+"'.")
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(val)
		case chunk.OP_DEFINE_GLOBAL:
			name := vm.readConstant().AsString()
			vm.Globals[name] = vm.pop()
		case chunk.OP_SET_GLOBAL:
			name := vm.readConstant().AsString()
			_, ok := vm.Globals[name]
			if !ok {
				loxerror.Error(-1, "Undefined variable '"+name+"'.")
			}
			vm.Globals[name] = vm.peek(0)
		case chunk.OP_EQUAL:
			b, a := vm.pop(), vm.pop()
			vm.push(value.BoolVal(a.Equals(b)))
		case chunk.OP_GREATER, chunk.OP_LESS,
			chunk.OP_ADD, chunk.OP_SUBTRACT, chunk.OP_MULTIPLY, chunk.OP_DIVIDE:
			vm.binaryOp(instruction)
		case chunk.OP_NOT:
			vm.push(value.BoolVal(vm.isFalsey(vm.pop())))
		case chunk.OP_NEGATE:
			if !vm.peek(0).IsNumber() {
				loxerror.Error(-1, "Operand must be a number.")
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(value.NumberVal(-vm.pop().AsNumber()))
		case chunk.OP_PRINT:
			printVal := vm.pop()
			fmt.Println(printVal.Data)
			vm.Out = printVal.Data
		case chunk.OP_RETURN:
			return INTERPRET_OK
		}
	}
}
