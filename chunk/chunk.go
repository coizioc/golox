package chunk

import (
	"fmt"
	"golox/value"
	"strings"
)

const (
	OP_CONSTANT byte = iota
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NOT
	OP_NEGATE
	OP_RETURN
)

type Chunk struct {
	Code      []byte
	Constants []value.Value
}

func New() *Chunk {
	return &Chunk{[]byte{}, []value.Value{}}
}

func (c *Chunk) String() string {
	sb := strings.Builder{}
	sb.WriteString("Constants: [")
	for _, constant := range c.Constants {
		sb.WriteString(fmt.Sprintf("%s ", constant.String()))
	}

	sb.WriteString("]\n\n")
	i := 0
	for i < len(c.Code) {
		sb.WriteString("\t")
		switch c.Code[i] {
		case OP_CONSTANT:
			i++
			constant := c.Constants[c.Code[i]]
			sb.WriteString(fmt.Sprintf("CONSTANT %v\n", constant))
		case OP_NIL:
			sb.WriteString("NIL\n")
		case OP_TRUE:
			sb.WriteString("TRUE\n")
		case OP_FALSE:
			sb.WriteString("FALSE\n")
		case OP_EQUAL:
			sb.WriteString("EQUAL\n")
		case OP_GREATER:
			sb.WriteString("GREATER\n")
		case OP_LESS:
			sb.WriteString("LESS\n")
		case OP_ADD:
			sb.WriteString("ADD\n")
		case OP_SUBTRACT:
			sb.WriteString("SUBTRACT\n")
		case OP_MULTIPLY:
			sb.WriteString("MULTIPLY\n")
		case OP_DIVIDE:
			sb.WriteString("DIVIDE\n")
		case OP_NOT:
			sb.WriteString("NOT\n")
		case OP_NEGATE:
			sb.WriteString("NEGATE\n")
		case OP_RETURN:
			sb.WriteString("RETURN\n")
		default:
			sb.WriteString(fmt.Sprintf("UNKNOWN_OP %v\n", c.Code[i]))
		}
		i++
	}
	return sb.String()
}

func (c *Chunk) Write(byte byte) {
	c.Code = append(c.Code, byte)
}

func (c *Chunk) AddValue(v value.Value) byte {
	c.Constants = append(c.Constants, v)
	return byte(len(c.Constants) - 1)
}
