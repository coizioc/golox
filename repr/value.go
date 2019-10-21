package repr

import (
	"fmt"
	"math"
)

type Type int

const (
	VAL_BOOL Type = iota
	VAL_NIL
	VAL_NUMBER
	VAL_STRING
	VAL_FUNCTION
	VAL_NATIVE
)

type Function struct {
	Chunk *Chunk
	Arity int
	Name  string
}

func (f *Function) String() string {
	if f.Name == "" {
		return "<script>"
	} else {
		return fmt.Sprintf("<fn %s>", f.Name)
	}
}

/*
func (f *Function) String() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Fn %s(argc: %d)\n\n", f.Name, f.Arity))
	sb.WriteString(f.Chunk.String())

	return sb.String()
} */

type FuncType int

const (
	FUNC_FUNCTION FuncType = iota
	FUNC_SCRIPT
)

type NativeFn func(argCount int, args []Value) Value

func (_ NativeFn) String() string {
	return "<native fn>"
}

type Value struct {
	Type Type
	Data interface{}
}

func BoolVal(value bool) Value {
	return Value{VAL_BOOL, value}
}

func NilVal() Value {
	return Value{VAL_NIL, nil}
}

func NumberVal(value float64) Value {
	return Value{VAL_NUMBER, value}
}

func StringVal(value string) Value {
	return Value{VAL_STRING, value}
}

func FunctionVal(value *Function) Value {
	return Value{VAL_FUNCTION, value}
}

func NativeVal(value NativeFn) Value {
	return Value{VAL_NATIVE, value}
}

func (v Value) AsBool() bool {
	return v.Data.(bool)
}

func (v Value) AsNumber() float64 {
	return v.Data.(float64)
}

func (v Value) AsString() string {
	return v.Data.(string)
}

func (v Value) AsFunction() *Function {
	return v.Data.(*Function)
}

func (v Value) AsNative() NativeFn {
	return v.Data.(NativeFn)
}

func (v Value) Equals(v2 Value) bool {
	if v.Type != v2.Type {
		return false
	}

	switch v.Type {
	case VAL_BOOL:
		return v.AsBool() == v2.AsBool()
	case VAL_NIL:
		return true
	case VAL_NUMBER:
		return v.AsNumber() == v2.AsNumber()
	case VAL_STRING:
		return v.AsString() == v2.AsString()
	case VAL_FUNCTION:
		return v.AsFunction().Chunk == v.AsFunction().Chunk
	case VAL_NATIVE:
		// TODO == NativeFn
		return false
	default:
		// Not reachable
		return false
	}
}

func (v Value) IsBool() bool {
	return v.Type == VAL_BOOL
}

func (v Value) IsNil() bool {
	return v.Type == VAL_NIL
}

func (v Value) IsNumber() bool {
	return v.Type == VAL_NUMBER
}

func (v Value) IsString() bool {
	return v.Type == VAL_STRING
}

func (v Value) IsFunction() bool {
	return v.Type == VAL_FUNCTION
}

func (v Value) IsNative() bool {
	return v.Type == VAL_NATIVE
}

func (v Value) String() string {
	switch v.Type {
	case VAL_BOOL:
		return fmt.Sprintf("%t", v.Data.(bool))
	case VAL_NIL:
		return fmt.Sprintf("nil")
	case VAL_NUMBER:
		num := v.Data.(float64)
		if math.Mod(num, 1.0) == 0 {
			return fmt.Sprintf("%.0f", num)
		}
		return fmt.Sprintf("%f", v.Data.(float64))
	case VAL_STRING:
		return fmt.Sprintf("%s", v.Data.(string))
	case VAL_FUNCTION:
		funcVal := v.AsFunction()
		if funcVal.Name == "" {
			return "<script>"
		} else {
			return fmt.Sprintf("<fn %s>", funcVal.Name)
		}
	case VAL_NATIVE:
		return "<native fn>"
	default:
		return fmt.Sprintf("%v", v.Data)
	}
}
