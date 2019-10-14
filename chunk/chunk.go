package chunk

import (
	"fmt"
	"strings"
)

const (
	OP_CONSTANT byte = iota
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NEGATE
	OP_RETURN
)

type Value float64

type Chunk struct {
	Code      []byte
	Constants []Value
}

func New() *Chunk {
	return &Chunk{[]byte{}, []Value{}}
}

func (c *Chunk) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%v\n\n", c.Constants))
	i := 0
	for i < len(c.Code) {
		sb.WriteString("\t")
		switch c.Code[i] {
		case OP_CONSTANT:
			i++
			constant := c.Constants[c.Code[i]]
			sb.WriteString(fmt.Sprintf("CONSTANT %v\n", constant))
		case OP_ADD:
			sb.WriteString("ADD\n")
		case OP_SUBTRACT:
			sb.WriteString("SUBTRACT\n")
		case OP_MULTIPLY:
			sb.WriteString("MULTIPLY\n")
		case OP_DIVIDE:
			sb.WriteString("DIVIDE\n")
		case OP_NEGATE:
			sb.WriteString("NEGATE\n")
		case OP_RETURN:
			sb.WriteString("RETURN\n")
		}
		i++
	}
	return sb.String()
}

func (c *Chunk) Write(byte byte) {
	c.Code = append(c.Code, byte)
}

func (c *Chunk) AddValue(value Value) byte {
	c.Constants = append(c.Constants, value)
	return byte(len(c.Constants) - 1)
}
