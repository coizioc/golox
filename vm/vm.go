package vm

import (
	"fmt"
	"golox/loxerror"
	"golox/parser"
	"golox/repr"
)

type InterpretResult byte

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Frames  []*CallFrame
	Stack   []repr.Value
	Globals map[string]repr.Value
	Out     interface{}
}

func New() *VM {
	return &VM{
		[]*CallFrame{},
		[]repr.Value{},
		make(map[string]repr.Value),
		nil,
	}
}

func (vm *VM) Interpret(source string) InterpretResult {
	p := parser.New(source)

	mainFunc := p.Compile()

	if mainFunc == nil {
		return INTERPRET_COMPILE_ERROR
	}
	vm.initNatives()

	funcValue := repr.FunctionVal(mainFunc)
	vm.push(funcValue)
	vm.callValue(funcValue, 0)
	return vm.run()
}

func (vm *VM) FrameCount() int {
	return len(vm.Frames)
}

func (vm *VM) CurrFrame() *CallFrame {
	return vm.Frames[vm.FrameCount()-1]
}

/* TODO add runtimeError handling.
func (vm *VM) runtimeError(msg string) {
	for i := vm.FrameCount() - 1; i >= 0; i-- {
		frame := vm.Frames[i]
		frameFunc := frame.Function
		if frameFunc.Name == "" {
			fmt.Fprintf(os.Stderr, "script\n")
		} else {
			fmt.Fprintf(os.Stderr, "%s()\n", frameFunc.Name)
		}
	}
}
*/

func (vm *VM) push(value repr.Value) {
	vm.Stack = append(vm.Stack, value)
}

func (vm *VM) pop() repr.Value {
	retValue := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return retValue
}

func (vm *VM) peek(distance int) repr.Value {
	return vm.Stack[len(vm.Stack)-1-distance]
}

func (vm *VM) call(calledFunc *repr.Function, argCount int) bool {
	if argCount != calledFunc.Arity {
		loxerror.Error(-1, fmt.Sprintf("Expected %d arguments but got %d.", calledFunc.Arity, argCount))
		return false
	}
	vm.AddFrame(calledFunc, 0, len(vm.Stack)-argCount-1)

	return true
}

func (vm *VM) callValue(callee repr.Value, argCount int) bool {
	if callee.IsFunction() {
		return vm.call(callee.AsFunction(), argCount)
	} else if callee.IsNative() {
		native := callee.AsNative()
		result := native(argCount, vm.Stack[len(vm.Stack)-argCount-1:])
		vm.push(result)
		return true
	} else {
		loxerror.Error(-1, "Can only call functions and classes.")
		return false
	}
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
	if op == repr.OP_ADD && bytea.IsString() && byteb.IsString() {
		vm.concatenate()
		return
	} else if bytea.IsNumber() && byteb.IsNumber() {
		b, a = vm.pop().AsNumber(), vm.pop().AsNumber()
	} else {
		loxerror.Error(-1, fmt.Sprintf("Operands must be a number [%v %d %v]\n", bytea, op, byteb))
	}

	switch op {
	case repr.OP_GREATER:
		vm.push(repr.BoolVal(a > b))
	case repr.OP_LESS:
		vm.push(repr.BoolVal(a < b))
	case repr.OP_ADD:
		vm.push(repr.NumberVal(a + b))
	case repr.OP_SUBTRACT:
		vm.push(repr.NumberVal(a - b))
	case repr.OP_MULTIPLY:
		vm.push(repr.NumberVal(a * b))
	case repr.OP_DIVIDE:
		vm.push(repr.NumberVal(a / b))
	}
}

func (vm *VM) concatenate() {
	b, a := vm.pop().AsString(), vm.pop().AsString()

	vm.push(repr.StringVal(a + b))
}

func (vm *VM) isFalsey(v repr.Value) bool {
	return v.IsNil() || (v.IsBool() && !v.AsBool())
}

func (vm *VM) readByte() byte {
	frame := vm.CurrFrame()
	byteRead := frame.Function.Chunk.Code[frame.IP]
	frame.IP++
	return byteRead
}

func (vm *VM) readConstant() repr.Value {
	byteRead := vm.readByte()
	return vm.CurrFrame().Function.Chunk.Constants[byteRead]
}

func (vm *VM) readShort() int {
	frame := vm.CurrFrame()
	frame.IP += 2
	short := int(frame.Function.Chunk.Code[frame.IP-2])<<8 | int(frame.Function.Chunk.Code[frame.IP-1])
	return short
}

func (vm *VM) run() InterpretResult {
	for {
		instruction := vm.readByte()
		//fmt.Printf("%d: Stack: %v\n", instruction, vm.Stack)
		switch instruction {
		case repr.OP_CONSTANT:
			constant := vm.readConstant()
			vm.push(constant)
		case repr.OP_NIL:
			vm.push(repr.NilVal())
		case repr.OP_TRUE:
			vm.push(repr.BoolVal(true))
		case repr.OP_FALSE:
			vm.push(repr.BoolVal(false))
		case repr.OP_POP:
			vm.pop()
		case repr.OP_GET_LOCAL:
			slot := int(vm.readByte())
			vm.push(vm.Stack[slot+vm.CurrFrame().StackStart])
		case repr.OP_SET_LOCAL:
			slot := int(vm.readByte())
			vm.Stack[slot+vm.CurrFrame().StackStart] = vm.peek(0)
		case repr.OP_GET_GLOBAL:
			name := vm.readConstant().AsString()
			val, ok := vm.Globals[name]
			if !ok {
				loxerror.Error(-1, fmt.Sprintf("Undefined variable '%s'.", name))
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(val)
		case repr.OP_DEFINE_GLOBAL:
			name := vm.readConstant().AsString()
			vm.Globals[name] = vm.pop()
		case repr.OP_SET_GLOBAL:
			name := vm.readConstant().AsString()
			_, ok := vm.Globals[name]
			if !ok {
				loxerror.Error(-1, fmt.Sprintf("Undefined variable '%s'.", name))
			}
			vm.Globals[name] = vm.peek(0)
		case repr.OP_EQUAL:
			b, a := vm.pop(), vm.pop()
			vm.push(repr.BoolVal(a.Equals(b)))
		case repr.OP_GREATER, repr.OP_LESS,
			repr.OP_ADD, repr.OP_SUBTRACT, repr.OP_MULTIPLY, repr.OP_DIVIDE:
			vm.binaryOp(instruction)
		case repr.OP_NOT:
			vm.push(repr.BoolVal(vm.isFalsey(vm.pop())))
		case repr.OP_NEGATE:
			if !vm.peek(0).IsNumber() {
				loxerror.Error(-1, "Operand must be a number.")
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(repr.NumberVal(-vm.pop().AsNumber()))
		case repr.OP_PRINT:
			printVal := vm.pop()
			fmt.Println(printVal.String())
			vm.Out = printVal.Data
		case repr.OP_JUMP:
			offset := vm.readShort()
			vm.CurrFrame().IP += offset
		case repr.OP_JUMP_IF_FALSE:
			offset := vm.readShort()
			if vm.isFalsey(vm.peek(0)) {
				vm.CurrFrame().IP += offset
			}
		case repr.OP_LOOP:
			offset := vm.readShort()
			vm.CurrFrame().IP -= offset
		case repr.OP_CALL:
			argCount := int(vm.readByte())
			if !vm.callValue(vm.peek(argCount), argCount) {
				return INTERPRET_RUNTIME_ERROR
			}
		case repr.OP_RETURN:
			result := vm.pop()
			vm.RemoveFrame()
			if vm.FrameCount() == 0 {
				return INTERPRET_OK
			}
			vm.push(result)
		}
	}
}
