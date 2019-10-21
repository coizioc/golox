package parser

import (
	"golox/token"
)

type Local struct {
	Name  token.Token
	Depth int
}

func (p *Parser) addLocal(name token.Token) {
	local := Local{name, -1}
	p.Compiler.Locals = append(p.Compiler.Locals, local)
}
