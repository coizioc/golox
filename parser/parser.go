package parser

import (
	"golox/chunk"
	"golox/loxerror"
	"golox/scanner"
	"golox/token"
	"golox/value"
)

const (
	PREC_NONE       = iota
	PREC_ASSIGNMENT // =
	PREC_OR         // or
	PREC_AND        // and
	PREC_EQUALITY   // == !==
	PREC_COMPARISON // < > <= >=
	PREC_TERM       // + -
	PREC_FACTOR     // * /
	PREC_UNARY      // ! -
	PREC_CALL       // . () []
	PREC_PRIMARY
)

type ParseFn = func()

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence int
}

var rules = make(map[token.Type]*ParseRule)

type Parser struct {
	Current        int
	Scanner        *scanner.Scanner
	CompilingChunk *chunk.Chunk
}

func New(source string, c *chunk.Chunk) *Parser {
	sc := scanner.New(source)

	return &Parser{0, sc, c}
}

func (p *Parser) Compile() bool {
	p.Scanner.ScanTokens()
	if loxerror.HadError {
		return false
	}
	if p.Scanner.Tokens[0].Type == token.EOF {
		return false
	}

	p.parse()

	p.endCompiler()

	return !loxerror.HadError
}

func (p *Parser) InitRules() {
	rules[token.LEFT_PAREN] = &ParseRule{p.grouping, nil, PREC_NONE}
	rules[token.RIGHT_PAREN] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.LEFT_BRACE] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.RIGHT_BRACE] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.COMMA] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.DOT] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.MINUS] = &ParseRule{p.unary, p.binary, PREC_TERM}
	rules[token.PLUS] = &ParseRule{nil, p.binary, PREC_TERM}
	rules[token.SEMICOLON] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.SLASH] = &ParseRule{nil, p.binary, PREC_FACTOR}
	rules[token.STAR] = &ParseRule{nil, p.binary, PREC_FACTOR}

	rules[token.BANG] = &ParseRule{p.unary, nil, PREC_NONE}
	rules[token.BANG_EQUAL] = &ParseRule{nil, p.binary, PREC_EQUALITY}
	rules[token.EQUAL] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.EQUAL_EQUAL] = &ParseRule{nil, p.binary, PREC_EQUALITY}
	rules[token.GREATER] = &ParseRule{nil, p.binary, PREC_COMPARISON}
	rules[token.GREATER_EQUAL] = &ParseRule{nil, p.binary, PREC_COMPARISON}
	rules[token.LESS] = &ParseRule{nil, p.binary, PREC_COMPARISON}
	rules[token.LESS_EQUAL] = &ParseRule{nil, p.binary, PREC_COMPARISON}

	rules[token.IDENTIFIER] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.STRING] = &ParseRule{p.string, nil, PREC_NONE}
	rules[token.NUMBER] = &ParseRule{p.number, nil, PREC_NONE}

	rules[token.AND] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.CLASS] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.ELSE] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.FALSE] = &ParseRule{p.literal, nil, PREC_NONE}
	rules[token.FOR] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.FUN] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.IF] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.NIL] = &ParseRule{p.literal, nil, PREC_NONE}
	rules[token.OR] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.PRINT] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.RETURN] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.SUPER] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.THIS] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.TRUE] = &ParseRule{p.literal, nil, PREC_NONE}
	rules[token.VAR] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.WHILE] = &ParseRule{nil, nil, PREC_NONE}

	rules[token.ILLEGAL] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.EOF] = &ParseRule{nil, nil, PREC_NONE}
}

func (p *Parser) CurrChunk() *chunk.Chunk {
	return p.CompilingChunk
}

func (p *Parser) CurrToken() token.Token {
	return p.Scanner.Tokens[p.Current]
}

func (p *Parser) PrevToken() token.Token {
	return p.Scanner.Tokens[p.Current-1]
}

func (p *Parser) advance() {
	p.Current += 1
}

func (p *Parser) endCompiler() {
	p.emitReturn()
}

func (p *Parser) emitByte(byte byte) {
	p.CurrChunk().Write(byte)
}

func (p *Parser) emitBytes(byte1, byte2 byte) {
	p.emitByte(byte1)
	p.emitByte(byte2)
}

func (p *Parser) emitConstant(value value.Value) {
	p.emitBytes(chunk.OP_CONSTANT, p.makeConstant(value))
}

func (p *Parser) emitReturn() {
	p.emitByte(chunk.OP_RETURN)
}

func (p *Parser) makeConstant(value value.Value) byte {
	constant := p.CurrChunk().AddValue(value)
	// TODO if constant > UINT8_MAX
	return constant
}

func (p *Parser) parse() {
	p.InitRules()
	p.expression()
	p.consume(token.EOF, "Expect end of expression.")
}

func (p *Parser) parsePrecedence(precedence int) {
	p.advance()
	prefixRule := p.getRule(p.PrevToken().Type).Prefix
	if prefixRule == nil {
		loxerror.Error(p.PrevToken().Line, "Expect expression.")
	} else {
		prefixRule()
	}

	for precedence <= p.getRule(p.CurrToken().Type).Precedence {
		p.advance()
		infixRule := p.getRule(p.PrevToken().Type).Infix
		infixRule()
	}
}

func (p *Parser) getRule(tokenType token.Type) *ParseRule {
	return rules[tokenType]
}

func (p *Parser) consume(tokenType token.Type, errMsg string) {
	if p.CurrToken().Type == tokenType {
		p.advance()
	} else {
		loxerror.Error(p.CurrToken().Line, errMsg)
	}
}

func (p *Parser) binary() {
	operatorType := p.PrevToken().Type

	rule := p.getRule(operatorType)
	p.parsePrecedence(rule.Precedence + 1)

	switch operatorType {
	case token.BANG_EQUAL:
		p.emitBytes(chunk.OP_EQUAL, chunk.OP_NOT)
	case token.EQUAL_EQUAL:
		p.emitByte(chunk.OP_EQUAL)
	case token.GREATER:
		p.emitByte(chunk.OP_GREATER)
	case token.GREATER_EQUAL:
		p.emitBytes(chunk.OP_LESS, chunk.OP_NOT)
	case token.LESS:
		p.emitByte(chunk.OP_LESS)
	case token.LESS_EQUAL:
		p.emitBytes(chunk.OP_GREATER, chunk.OP_NOT)
	case token.PLUS:
		p.emitByte(chunk.OP_ADD)
	case token.MINUS:
		p.emitByte(chunk.OP_SUBTRACT)
	case token.STAR:
		p.emitByte(chunk.OP_MULTIPLY)
	case token.SLASH:
		p.emitByte(chunk.OP_DIVIDE)
	}
}

func (p *Parser) expression() {
	p.parsePrecedence(PREC_ASSIGNMENT)
}

func (p *Parser) grouping() {
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after expression.")
}

func (p *Parser) literal() {
	// fmt.Printf("Prev token: %s", p.Scanner.Tokens[p.Current - 1].String())
	switch p.PrevToken().Type {
	case token.FALSE:
		p.emitByte(chunk.OP_FALSE)
	case token.NIL:
		p.emitByte(chunk.OP_NIL)
	case token.TRUE:
		p.emitByte(chunk.OP_TRUE)
	default:
		// Unreachable
		return
	}
}

func (p *Parser) number() {
	val := p.PrevToken().Literal.(float64)
	p.emitConstant(value.NumberVal(val))
}

func (p *Parser) string() {
	val := p.PrevToken().Literal.(string)
	p.emitConstant(value.StringVal(val))
}

func (p *Parser) unary() {
	operatorType := p.PrevToken().Type
	p.parsePrecedence(PREC_UNARY)
	switch operatorType {
	case token.BANG:
		p.emitByte(chunk.OP_NOT)
	case token.MINUS:
		p.emitByte(chunk.OP_NEGATE)
	default:
		return
	}
}
