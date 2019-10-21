package parser

import (
	"golox/loxerror"
	"golox/repr"
	"golox/scanner"
	"golox/token"
)

type Parser struct {
	Current  int
	Scanner  *scanner.Scanner
	Compiler *Compiler
}

func New(source string) *Parser {
	sc := scanner.New(source)
	comp := InitCompiler(repr.FUNC_SCRIPT, "")

	return &Parser{0, sc, comp}
}

func (p *Parser) Compile() *repr.Function {
	p.Scanner.ScanTokens()
	if loxerror.HadError {
		return nil
	}
	if p.Scanner.Tokens[0].Type == token.EOF {
		return nil
	}

	p.parse()

	if loxerror.HadError {
		return nil
	} else {
		return p.endCompiler()
	}
}

func (p *Parser) CurrChunk() *repr.Chunk {
	return p.Compiler.Function.Chunk
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

func (p *Parser) emitByte(byte byte) {
	p.CurrChunk().Write(byte)
}

func (p *Parser) emitBytes(byte1, byte2 byte) {
	p.emitByte(byte1)
	p.emitByte(byte2)
}

func (p *Parser) emitConstant(value repr.Value) {
	p.emitBytes(repr.OP_CONSTANT, p.makeConstant(value))
}

func (p *Parser) emitJump(instruction byte) int {
	p.emitByte(instruction)
	// Store jump address as 16-bit number.
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.CurrChunk().Code) - 2
}

func (p *Parser) emitLoop(loopStart int) {
	p.emitByte(repr.OP_LOOP)

	offset := len(p.CurrChunk().Code) - loopStart + 2
	if offset > 65535 {
		loxerror.Error(-1, "Loop body too large.")
	}

	p.emitByte(byte(offset>>8) & 0xff)
	p.emitByte(byte(offset) & 0xff)
}

func (p *Parser) emitReturn() {
	p.emitByte(repr.OP_NIL)
	p.emitByte(repr.OP_RETURN)
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

func (p *Parser) makeConstant(value repr.Value) byte {
	constant := p.CurrChunk().AddValue(value)
	// TODO if constant > UINT8_MAX
	return constant
}

func (p *Parser) parse() {
	p.InitRules()
	for !p.match(token.EOF) {
		p.declaration()
	}
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

func (p *Parser) declareVariable() {
	if p.Compiler.ScopeDepth == 0 {
		return
	}

	name := p.PrevToken()

	for i := len(p.Compiler.Locals) - 1; i >= 0; i-- {
		local := p.Compiler.Locals[i]
		if local.Depth != -1 && local.Depth < p.Compiler.ScopeDepth {
			break
		}

		if p.identifiersEqual(name, local.Name) {
			loxerror.Error(-1, "Variable with this name already declared in this scope.")
		}
	}

	p.addLocal(name)
}

func (p *Parser) defineVariable(global byte) {
	if p.Compiler.ScopeDepth > 0 {
		p.markInitialized()
		return
	}

	p.emitBytes(repr.OP_DEFINE_GLOBAL, global)
}

func (p *Parser) namedVariable(name token.Token, canAssign bool) {
	var arg, getOp, setOp byte
	res := p.resolveLocal(name)
	if res != -1 {
		arg = byte(res)
		getOp = repr.OP_GET_LOCAL
		setOp = repr.OP_SET_LOCAL
	} else {
		arg = p.identifierConstant(name)
		getOp = repr.OP_GET_GLOBAL
		setOp = repr.OP_SET_GLOBAL
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
	if p.Compiler.ScopeDepth > 0 {
		return 0
	}

	return p.identifierConstant(p.PrevToken())
}

func (p *Parser) identifierConstant(name token.Token) byte {
	return p.CurrChunk().AddValue(repr.StringVal(name.Lexeme))
}

func (p *Parser) identifiersEqual(a, b token.Token) bool {
	return a.Lexeme == b.Lexeme
}

func (p *Parser) markInitialized() {
	if p.Compiler.ScopeDepth == 0 {
		return
	}
	p.Compiler.Locals[len(p.Compiler.Locals)-1].Depth = p.Compiler.ScopeDepth
}

func (p *Parser) resolveLocal(name token.Token) int {
	for i := len(p.Compiler.Locals) - 1; i >= 0; i-- {
		local := p.Compiler.Locals[i]
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
	p.Compiler.ScopeDepth++
}

func (p *Parser) endScope() {
	p.Compiler.ScopeDepth--

	for len(p.Compiler.Locals) > 0 && p.Compiler.Locals[len(p.Compiler.Locals)-1].Depth > p.Compiler.ScopeDepth {
		p.emitByte(repr.OP_POP)
		p.Compiler.Locals = p.Compiler.Locals[:len(p.Compiler.Locals)-1]
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
	endJump := p.emitJump(repr.OP_JUMP_IF_FALSE)
	p.emitByte(repr.OP_POP)
	p.parsePrecedence(PREC_AND)
	p.patchJump(endJump)
}

func (p *Parser) argumentList() byte {
	argCount := 0
	if !p.check(token.RIGHT_PAREN) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			p.expression()

			if argCount == 255 {
				loxerror.Error(p.CurrToken().Line, "Cannot have more than 255 arguments.")
			}
			argCount++
		}
	}

	p.consume(token.RIGHT_PAREN, "Expect ')' after arguments.")
	return byte(argCount)
}

func (p *Parser) binary(canAssign bool) {
	operatorType := p.PrevToken().Type

	rule := p.getRule(operatorType)
	p.parsePrecedence(rule.Precedence + 1)

	switch operatorType {
	case token.BANG_EQUAL:
		p.emitBytes(repr.OP_EQUAL, repr.OP_NOT)
	case token.EQUAL_EQUAL:
		p.emitByte(repr.OP_EQUAL)
	case token.GREATER:
		p.emitByte(repr.OP_GREATER)
	case token.GREATER_EQUAL:
		p.emitBytes(repr.OP_LESS, repr.OP_NOT)
	case token.LESS:
		p.emitByte(repr.OP_LESS)
	case token.LESS_EQUAL:
		p.emitBytes(repr.OP_GREATER, repr.OP_NOT)
	case token.PLUS:
		p.emitByte(repr.OP_ADD)
	case token.MINUS:
		p.emitByte(repr.OP_SUBTRACT)
	case token.STAR:
		p.emitByte(repr.OP_MULTIPLY)
	case token.SLASH:
		p.emitByte(repr.OP_DIVIDE)
	}
}

func (p *Parser) block() {
	for !p.check(token.RIGHT_BRACE) && !p.check(token.EOF) {
		p.declaration()
	}
	p.consume(token.RIGHT_BRACE, "Expect '}' after block.")
}

func (p *Parser) call(canAssign bool) {
	argCount := p.argumentList()
	p.emitBytes(repr.OP_CALL, argCount)
}

func (p *Parser) declaration() {
	if p.match(token.VAR) {
		p.varDeclaration()
	} else if p.match(token.FUN) {
		p.funDeclaration()
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
	p.emitByte(repr.OP_POP)
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

		exitJump = p.emitJump(repr.OP_JUMP_IF_FALSE)
		p.emitByte(repr.OP_POP)
	}

	if !p.match(token.RIGHT_PAREN) {
		bodyJump := p.emitJump(repr.OP_JUMP)

		incrementStart := len(p.CurrChunk().Code)
		p.expression()
		p.emitByte(repr.OP_POP)
		p.consume(token.RIGHT_PAREN, "Expect ')' after for clauses.")

		p.emitLoop(loopStart)
		loopStart = incrementStart
		p.patchJump(bodyJump)
	}

	p.statement()

	p.emitLoop(loopStart)

	if exitJump != -1 {
		p.patchJump(exitJump)
		p.emitByte(repr.OP_POP)
	}

	p.endScope()
}

func (p *Parser) function(funcType repr.FuncType) {
	funcCompiler := InitCompiler(funcType, p.PrevToken().Lexeme)
	p.encloseCompiler(funcCompiler)
	p.beginScope()

	p.consume(token.LEFT_PAREN, "Expect '(' after function name.")
	if !p.check(token.RIGHT_PAREN) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			p.Compiler.Function.Arity++
			if p.Compiler.Function.Arity > 255 {
				loxerror.Error(p.CurrToken().Line, "Cannot have more than 255 parameters.")
			}

			paramConstant := p.parseVariable("Expect parameter name.")
			p.defineVariable(paramConstant)
		}
	}
	p.consume(token.RIGHT_PAREN, "Expect ')' after parameters.")

	p.consume(token.LEFT_BRACE, "Expect '{' before function body.")
	p.block()

	compiledFunction := p.endCompiler()

	p.restoreCompiler()

	p.emitBytes(repr.OP_CONSTANT, p.makeConstant(repr.FunctionVal(compiledFunction)))
}

func (p *Parser) funDeclaration() {
	global := p.parseVariable("Expect function name.")
	p.markInitialized()
	p.function(repr.FUNC_FUNCTION)
	p.defineVariable(global)
}

func (p *Parser) grouping(canAssign bool) {
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after expression.")
}

func (p *Parser) ifStatement() {
	p.consume(token.LEFT_PAREN, "Expect '(' after 'if'.")
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after condition.")

	thenJump := p.emitJump(repr.OP_JUMP_IF_FALSE)
	p.emitByte(repr.OP_POP)
	p.statement()

	elseJump := p.emitJump(repr.OP_JUMP)

	p.patchJump(thenJump)
	p.emitByte(repr.OP_POP)

	if p.match(token.ELSE) {
		p.statement()
	}
	p.patchJump(elseJump)
}

func (p *Parser) literal(canAssign bool) {
	// fmt.Printf("Prev token: %s", p.Scanner.Tokens[p.Current - 1].ValStr())
	switch p.PrevToken().Type {
	case token.FALSE:
		p.emitByte(repr.OP_FALSE)
	case token.NIL:
		p.emitByte(repr.OP_NIL)
	case token.TRUE:
		p.emitByte(repr.OP_TRUE)
	default:
		// Unreachable
		return
	}
}

func (p *Parser) number(canAssign bool) {
	val := p.PrevToken().Literal.(float64)
	p.emitConstant(repr.NumberVal(val))
}

func (p *Parser) or(canAssign bool) {
	elseJump := p.emitJump(repr.OP_JUMP_IF_FALSE)
	endJump := p.emitJump(repr.OP_JUMP)

	p.patchJump(elseJump)
	p.emitByte(repr.OP_POP)

	p.parsePrecedence(PREC_OR)
	p.patchJump(endJump)
}

func (p *Parser) printStatement() {
	p.expression()
	p.consume(token.SEMICOLON, "Expect ; after value.")
	p.emitByte(repr.OP_PRINT)
}

func (p *Parser) returnStatement() {
	if p.Compiler.Type == repr.FUNC_SCRIPT {
		loxerror.Error(p.CurrToken().Line, "Cannot return from top-level code.")
	}
	if p.match(token.SEMICOLON) {
		p.emitReturn()
	} else {
		p.expression()
		p.consume(token.SEMICOLON, "Expect ';' after return value.")
		p.emitByte(repr.OP_RETURN)
	}
}

func (p *Parser) statement() {
	if p.match(token.PRINT) {
		p.printStatement()
	} else if p.match(token.FOR) {
		p.forStatement()
	} else if p.match(token.IF) {
		p.ifStatement()
	} else if p.match(token.RETURN) {
		p.returnStatement()
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
	p.emitConstant(repr.StringVal(val))
}

func (p *Parser) unary(canAssign bool) {
	operatorType := p.PrevToken().Type
	p.parsePrecedence(PREC_UNARY)
	switch operatorType {
	case token.BANG:
		p.emitByte(repr.OP_NOT)
	case token.MINUS:
		p.emitByte(repr.OP_NEGATE)
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
		p.emitByte(repr.OP_NIL)
	}
	p.consume(token.SEMICOLON, "Expect ';' after variable declaration.")
	p.defineVariable(global)
}

func (p *Parser) whileStatement() {
	loopStart := len(p.CurrChunk().Code)

	p.consume(token.LEFT_PAREN, "Expect '(' after 'while'.")
	p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := p.emitJump(repr.OP_JUMP_IF_FALSE)

	p.emitByte(repr.OP_POP)
	p.statement()

	p.emitLoop(loopStart)

	p.patchJump(exitJump)
	p.emitByte(repr.OP_POP)
}
