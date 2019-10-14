package value

import (
	"fmt"
)

type Type int
type ObjType int

const (
	VAL_BOOL Type = iota
	VAL_NIL
	VAL_NUMBER
	VAL_STRING
)

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

func (v Value) AsBool() bool {
	return v.Data.(bool)
}

func (v Value) AsNumber() float64 {
	return v.Data.(float64)
}

func (v Value) AsString() string {
	return v.Data.(string)
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

func (v Value) String() string {
	switch v.Type {
	case VAL_BOOL:
		return fmt.Sprintf("%t", v.Data.(bool))
	case VAL_NIL:
		return fmt.Sprintf("nil")
	case VAL_NUMBER:
		return fmt.Sprintf("%f", v.Data.(float64))
	case VAL_STRING:
		return fmt.Sprintf("%s", v.Data.(string))
	default:
		return fmt.Sprintf("%v", v.Data)
	}
}
