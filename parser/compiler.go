package parser

import (
	"golox/repr"
	"golox/token"
)

type Compiler struct {
	Enclosing  *Compiler
	Function   *repr.Function
	Type       repr.FuncType
	Locals     []Local
	ScopeDepth int
}

func InitCompiler(funcType repr.FuncType, name string) *Compiler {
	return &Compiler{
		nil,
		&repr.Function{Chunk: repr.NewChunk(), Name: name},
		funcType,
		[]Local{{token.Token{}, 0}},
		0,
	}
}

func (p *Parser) endCompiler() *repr.Function {
	p.emitReturn()
	compiledFunc := p.Compiler.Function
	return compiledFunc
}

func (p *Parser) encloseCompiler(enclosingComp *Compiler) {
	enclosingComp.Enclosing = p.Compiler
	p.Compiler = enclosingComp
}

func (p *Parser) restoreCompiler() {
	p.Compiler = p.Compiler.Enclosing
}
