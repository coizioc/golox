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

type ParseFn = func(canAssign bool)

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence int
}

var rules = make(map[token.Type]*ParseRule)

type Local struct {
	Name  token.Token
	Depth int
}
type Parser struct {
	Locals         []Local
	ScopeDepth     int
	Current        int
	Scanner        *scanner.Scanner
	CompilingChunk *chunk.Chunk
}

func New(source string, c *chunk.Chunk) *Parser {
	sc := scanner.New(source)

	return &Parser{[]Local{}, 0, 0, sc, c}
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

	rules[token.IDENTIFIER] = &ParseRule{p.variable, nil, PREC_NONE}
	rules[token.STRING] = &ParseRule{p.string, nil, PREC_NONE}
	rules[token.NUMBER] = &ParseRule{p.number, nil, PREC_NONE}

	rules[token.AND] = &ParseRule{nil, p.and, PREC_AND}
	rules[token.CLASS] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.ELSE] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.FALSE] = &ParseRule{p.literal, nil, PREC_NONE}
	rules[token.FOR] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.FUN] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.IF] = &ParseRule{nil, nil, PREC_NONE}
	rules[token.NIL] = &ParseRule{p.literal, nil, PREC_NONE}
	rules[token.OR] = &ParseRule{nil, p.or, PREC_OR}
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

func (p *Parser) check(tokenType token.Type) bool {
	return p.CurrToken().Type == tokenType
}

func (p *Parser) match(tokenType token.Type) bool {
	if !p.check(tokenType) {
		return false
	}
	p.advance()
	return true
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

func (p *Parser) emitJump(instruction byte) int {
	p.emitByte(instruction)
	// Store jump address as 16-bit number.
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.CurrChunk().Code) - 2
}

func (p *Parser) emitLoop(loopStart int) {
	p.emitByte(chunk.OP_LOOP)

	offset := len(p.CurrChunk().Code) - loopStart + 2
	if offset > 65535 {
		loxerror.Error(-1, "Loop body too large.")
	}

	p.emitByte(byte(offset>>8) & 0xff)
	p.emitByte(byte(offset) & 0xff)
}

func (p *Parser) emitReturn() {
	p.emitByte(chunk.OP_RETURN)
}

func (p *Parser) patchJump(offset int) {
	jump := len(p.CurrChunk().Code) - offset - 2

	// If jump is greater than 2^16.
	if jump > 65536 {
		loxerror.Error(-1, "Too much code to jump over.")
	}

	p.CurrChunk().Code[offset] = byte((jump >> 8) & 0xff)
	p.CurrChunk().Code[offset+1] = byte(jump & 0xff)
}

func (p *Parser) makeConstant(value value.Value) byte {
	constant := p.CurrChunk().AddValue(value)
	// TODO if constant > UINT8_MAX
	return constant
}

func (p *Parser) parse() {
	p.InitRules()
	for !p.match(token.EOF) {
		p.declaration()
	}
	p.endCompiler()
}

func (p *Parser) parsePrecedence(precedence int) {
	p.advance()
	prefixRule := p.getRule(p.PrevToken().Type).Prefix
	if prefixRule == nil {
		loxerror.Error(p.PrevToken().Line, "Expect expression.")
		return
	}

	canAssign := precedence <= PREC_ASSIGNMENT
	prefixRule(canAssign)

	for precedence <= p.getRule(p.CurrToken().Type).Precedence {
		p.advance()
		infixRule := p.getRule(p.PrevToken().Type).Infix
		infixRule(canAssign)
	}

	if canAssign && p.match(token.EQUAL) {
		loxerror.Error(-1, "Invalid assignment target.")
		p.expression()
	}
}

func (p *Parser) getRule(tokenType token.Type) *ParseRule {
	return rules[tokenType]
}

func (p *Parser) addLocal(name token.Token) {
	local := Local{name, -1}
	p.Locals = append(p.Locals, local)
}

func (p *Parser) declareVariable() {
	if p.ScopeDepth == 0 {
		return
	}

	name := p.PrevToken()

	for i := len(p.Locals) - 1; i >= 0; i-- {
		local := p.Locals[i]
		if local.Depth != -1 && local.Depth < p.ScopeDepth {
			break
		}

		if p.identifiersEqual(name, local.Name) {
			loxerror.Error(-1, "Variable with this name already declared in this scope.")
		}
	}

	p.addLocal(name)
}

func (p *Parser) defineVariable(global byte) {
	if p.ScopeDepth > 0 {
		p.markInitialized()
		return
	}

	p.emitBytes(chunk.OP_DEFINE_GLOBAL, global)
}

func (p *Parser) namedVariable(name token.Token, canAssign bool) {
	var arg, getOp, setOp byte
	res := p.resolveLocal(name)
	if res != -1 {
		arg = byte(res)
		getOp = chunk.OP_GET_LOCAL
		setOp = chunk.OP_SET_LOCAL
	} else {
		arg = p.identifierConstant(name)
		getOp = chunk.OP_GET_GLOBAL
		setOp = chunk.OP_SET_GLOBAL
	}

	if canAssign && p.match(token.EQUAL) {
		p.expression()
		p.emitBytes(setOp, arg)
	} else {
		p.emitBytes(getOp, arg)
	}
}

func (p *Parser) parseVariable(errorMessage string) byte {
	p.consume(token.IDENTIFIER, errorMessage)

	p.declareVariable()
	if p.ScopeDepth > 0 {
		return 0
	}

	return p.identifierConstant(p.PrevToken())
}

func (p *Parser) identifierConstant(name token.Token) byte {
	return p.CurrChunk().AddValue(value.StringVal(name.Lexeme))
}

func (p *Parser) identifiersEqual(a, b token.Token) bool {
	return a.Lexeme == b.Lexeme
}

func (p *Parser) markInitialized() {
	p.Locals[len(p.Locals)-1].Depth = p.ScopeDepth
}

func (p *Parser) resolveLocal(name token.Token) int {
	for i := len(p.Locals) - 1; i >= 0; i-- {
		local := p.Locals[i]
		if p.identifiersEqual(name, local.Name) {
			if local.Depth == -1 {
				loxerror.Error(-1, "Cannot read local variable in its own initializer.")
			}
			return i
		}
	}

	return -1
}

func (p *Parser) beginScope() {
	p.ScopeDepth++
}

func (p *Parser) endScope() {
	p.ScopeDepth--

	for len(p.Locals) > 0 && p.Locals[len(p.Locals)-1].Depth > p.ScopeDepth {
		p.emitByte(chunk.OP_POP)
		p.Locals = p.Locals[:len(p.Locals)-1]
	}
}

func (p *Parser) consume(tokenType token.Type, errMsg string) {
	if p.CurrToken().Type == tokenType {
		p.advance()
	} else {
		loxerror.Error(p.CurrToken().Line, errMsg)
	}
}

func (p *Parser) and(canAssign bool) {
	endJump := p.emitJump(chunk.OP_JUMP_IF_FALSE)
	p.emitByte(chunk.OP_POP)
	p.parsePrecedence(PREC_AND)
	p.patchJump(endJump)
}

func (p *Parser) binary(canAssign bool) {
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

func (p *Parser) block() {
	for !p.check(token.RIGHT_BRACE) && !p.check(token.EOF) {
		p.declaration()
	}
	p.consume(token.RIGHT_BRACE, "Expect '}' after block.")
}

func (p *Parser) declaration() {
	if p.match(token.VAR) {
		p.varDeclaration()
	} else {
		p.statement()
	}
}

func (p *Parser) expression() {
	p.parsePrecedence(PREC_ASSIGNMENT)
}

func (p *Parser) expressionStatement() {
	p.expression()
	p.consume(token.SEMICOLON, "Expect ';' after expression.")
	p.emitByte(chunk.OP_POP)
}

func (p *Parser) forStatement() {
	p.beginScope()

	p.consume(token.LEFT_PAREN, "Expect '(' after 'for'.")
	if p.match(token.SEMICOLON) {
		// No initializer.
	} else if p.match(token.VAR) {
		p.varDeclaration()
	} else {
		p.expressionStatement()
	}

	loopStart := len(p.CurrChunk().Code)

	exitJump := -1
	if !p.match(token.SEMICOLON) {
		p.expression()
		p.consume(token.SEMICOLON, "Expect ';' after loop condition.")

		exitJump = p.emitJump(chunk.OP_JUMP_IF_FALSE)
		p.emitByte(chunk.OP_POP)
	}

	if !p.match(token.RIGHT_PAREN) {
		bodyJump := p.emitJump(chunk.OP_JUMP)

		incrementStart := len(p.CurrChunk().Code)
		p.expression()
		p.emitByte(chunk.OP_POP)
		p.consume(token.RIGHT_PAREN, "Expect ')' after for clauses.")

		p.emitLoop(loopStart)
		loopStart = incrementStart
		p.patchJump(bodyJump)
	}

	p.statement()

	p.emitLoop(loopStart)

	if exitJump != -1 {
		p.patchJump(exitJump)
		p.emitByte(chunk.OP_POP)
	}

	p.endScope()
}

func (p *Parser) grouping(canAssign bool) {
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after expression.")
}

func (p *Parser) ifStatement() {
	p.consume(token.LEFT_PAREN, "Expect '(' after 'if'.")
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after condition.")

	thenJump := p.emitJump(chunk.OP_JUMP_IF_FALSE)
	p.emitByte(chunk.OP_POP)
	p.statement()

	elseJump := p.emitJump(chunk.OP_JUMP)

	p.patchJump(thenJump)
	p.emitByte(chunk.OP_POP)

	if p.match(token.ELSE) {
		p.statement()
	}
	p.patchJump(elseJump)
}

func (p *Parser) literal(canAssign bool) {
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

func (p *Parser) number(canAssign bool) {
	val := p.PrevToken().Literal.(float64)
	p.emitConstant(value.NumberVal(val))
}

func (p *Parser) or(canAssign bool) {
	elseJump := p.emitJump(chunk.OP_JUMP_IF_FALSE)
	endJump := p.emitJump(chunk.OP_JUMP)

	p.patchJump(elseJump)
	p.emitByte(chunk.OP_POP)

	p.parsePrecedence(PREC_OR)
	p.patchJump(endJump)
}

func (p *Parser) printStatement() {
	p.expression()
	p.consume(token.SEMICOLON, "Expect ; after value.")
	p.emitByte(chunk.OP_PRINT)
}

func (p *Parser) statement() {
	if p.match(token.PRINT) {
		p.printStatement()
	} else if p.match(token.FOR) {
		p.forStatement()
	} else if p.match(token.IF) {
		p.ifStatement()
	} else if p.match(token.LEFT_BRACE) {
		p.beginScope()
		p.block()
		p.endScope()
	} else if p.match(token.WHILE) {
		p.whileStatement()
	} else {
		p.expressionStatement()
	}
}

func (p *Parser) string(canAssign bool) {
	val := p.PrevToken().Literal.(string)
	p.emitConstant(value.StringVal(val))
}

func (p *Parser) unary(canAssign bool) {
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

func (p *Parser) variable(canAssign bool) {
	p.namedVariable(p.PrevToken(), canAssign)
}

func (p *Parser) varDeclaration() {
	global := p.parseVariable("Expect variable name")
	if p.match(token.EQUAL) {
		p.expression()
	} else {
		p.emitByte(chunk.OP_NIL)
	}
	p.consume(token.SEMICOLON, "Expect ';' after variable declaration.")
	p.defineVariable(global)
}

func (p *Parser) whileStatement() {
	loopStart := len(p.CurrChunk().Code)

	p.consume(token.LEFT_PAREN, "Expect '(' after 'while'.")
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := p.emitJump(chunk.OP_JUMP_IF_FALSE)

	p.emitByte(chunk.OP_POP)
	p.statement()

	p.emitLoop(loopStart)

	p.patchJump(exitJump)
	p.emitByte(chunk.OP_POP)
}
