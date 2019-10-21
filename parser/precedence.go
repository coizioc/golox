package parser

import "golox/token"

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

func (p *Parser) InitRules() {
	rules[token.LEFT_PAREN] = &ParseRule{p.grouping, p.call, PREC_CALL}
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
