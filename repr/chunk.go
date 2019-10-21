package repr

import (
	"fmt"
	"strings"
)

const (
	OP_CONSTANT byte = iota
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_POP
	OP_GET_LOCAL
	OP_SET_LOCAL
	OP_GET_GLOBAL
	OP_DEFINE_GLOBAL
	OP_SET_GLOBAL
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NOT
	OP_NEGATE
	OP_PRINT
	OP_JUMP
	OP_JUMP_IF_FALSE
	OP_LOOP
	OP_CALL
	OP_RETURN
)

type Chunk struct {
	Code      []byte
	Constants []Value
}

func NewChunk() *Chunk {
	return &Chunk{[]byte{}, []Value{}}
}

func (c *Chunk) String() string {
	sb := strings.Builder{}
	sb.WriteString("Constants: [")
	for _, constant := range c.Constants {
		sb.WriteString(fmt.Sprintf("%s ", constant.String()))
	}

	sb.WriteString("]\n\n")
	ip := 0
	for ip < len(c.Code) {
		sb.WriteString(fmt.Sprintf("\t%3d ", c.Code[ip]))
		switch c.Code[ip] {
		case OP_CONSTANT:
			ip++
			constant := c.Constants[c.Code[ip]]
			sb.WriteString(fmt.Sprintf("CONSTANT %v\n", constant))
		case OP_NIL:
			sb.WriteString("NIL\n")
		case OP_TRUE:
			sb.WriteString("TRUE\n")
		case OP_FALSE:
			sb.WriteString("FALSE\n")
		case OP_POP:
			sb.WriteString("POP\n")
		case OP_GET_LOCAL:
			ip++
			//constant := c.Constants[c.Code[ip]]
			sb.WriteString(fmt.Sprintf("GET_LOCAL &%v\n", ip))
		case OP_SET_LOCAL:
			ip++
			//constant := c.Constants[c.Code[ip]]
			sb.WriteString(fmt.Sprintf("SET_LOCAL &%v\n", ip))
		case OP_GET_GLOBAL:
			ip++
			constant := c.Constants[c.Code[ip]]
			sb.WriteString(fmt.Sprintf("GET_GLOBAL %v\n", constant))
		case OP_DEFINE_GLOBAL:
			ip++
			constant := c.Constants[c.Code[ip]]
			sb.WriteString(fmt.Sprintf("DEFINE_GLOBAL %v\n", constant))
		case OP_SET_GLOBAL:
			ip++
			constant := c.Constants[c.Code[ip]]
			sb.WriteString(fmt.Sprintf("SET_GLOBAL %v\n", constant))
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
		case OP_PRINT:
			sb.WriteString("PRINT\n")
		case OP_JUMP:
			ip += 3
			jumpLen := int(c.Code[ip])<<8 | int(c.Code[ip-1])
			sb.WriteString(fmt.Sprintf("JUMP %d\n", jumpLen))
		case OP_JUMP_IF_FALSE:
			ip += 2
			jumpLen := int(c.Code[ip])<<8 | int(c.Code[ip-1])
			sb.WriteString(fmt.Sprintf("JUMP_IF_FALSE %d\n", jumpLen))
		case OP_LOOP:
			ip += 2
			jumpLen := int(c.Code[ip])<<8 | int(c.Code[ip-1])
			sb.WriteString(fmt.Sprintf("LOOP %d\n", jumpLen))
		case OP_CALL:
			ip++
			argCount := c.Code[ip]
			sb.WriteString(fmt.Sprintf("CALL %d\n", argCount))
		case OP_RETURN:
			sb.WriteString("RETURN\n")
		default:
			sb.WriteString(fmt.Sprintf("UNKNOWN_OP %v\n", c.Code[ip]))
		}
		ip++
	}
	return sb.String()
}

func (c *Chunk) Write(byte byte) {
	c.Code = append(c.Code, byte)
}

func (c *Chunk) AddValue(v Value) byte {
	c.Constants = append(c.Constants, v)
	return byte(len(c.Constants) - 1)
}
