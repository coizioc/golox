package tests

import (
	"golox/scanner"
	"golox/token"
	"testing"
)

func runScanner(t *testing.T, source string, expected []token.Token) {
	loxScanner := scanner.New(source)
	loxScanner.ScanTokens()
	for i := range loxScanner.Tokens {
		if expected[i] != loxScanner.Tokens[i] {
			t.Errorf("Unexpected token. Expected: %s. Got: %s.", expected[i].String(), loxScanner.Tokens[i].String())
		}
	}
}

func TestIdentifiers(t *testing.T) {
	source := `andy formless fo _ _123 _abc ab123
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_`

	expected := []token.Token{
		{token.IDENTIFIER, "andy", nil, 1},
		{token.IDENTIFIER, "formless", nil, 1},
		{token.IDENTIFIER, "fo", nil, 1},
		{token.IDENTIFIER, "_", nil, 1},
		{token.IDENTIFIER, "_123", nil, 1},
		{token.IDENTIFIER, "_abc", nil, 1},
		{token.IDENTIFIER, "ab123", nil, 1},
		{token.IDENTIFIER, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_", nil, 2},
		{token.EOF, "", nil, 2},
	}

	runScanner(t, source, expected)
}

func TestKeywords(t *testing.T) {
	source := "and class else false for fun if nil or return super this true var while"

	expected := []token.Token{
		{token.AND, "and", nil, 1},
		{token.CLASS, "class", nil, 1},
		{token.ELSE, "else", nil, 1},
		{token.FALSE, "false", nil, 1},
		{token.FOR, "for", nil, 1},
		{token.FUN, "fun", nil, 1},
		{token.IF, "if", nil, 1},
		{token.NIL, "nil", nil, 1},
		{token.OR, "or", nil, 1},
		{token.RETURN, "return", nil, 1},
		{token.SUPER, "super", nil, 1},
		{token.THIS, "this", nil, 1},
		{token.TRUE, "true", nil, 1},
		{token.VAR, "var", nil, 1},
		{token.WHILE, "while", nil, 1},
		{token.EOF, "", nil, 1},
	}

	runScanner(t, source, expected)
}

func TestNumbers(t *testing.T) {
	source := `123
123.456
.456
123.`

	expected := []token.Token{
		{token.NUMBER, "123", 123.0, 1},
		{token.NUMBER, "123.456", 123.456, 2},
		{token.DOT, ".", nil, 3},
		{token.NUMBER, "456", 456.0, 3},
		{token.NUMBER, "123", 123.0, 4},
		{token.DOT, ".", nil, 4},
		{token.EOF, "", nil, 4},
	}

	runScanner(t, source, expected)
}

func TestPunctuators(t *testing.T) {
	source := "(){};,+-*!===<=>=!=<>/."

	expected := []token.Token{
		{token.LEFT_PAREN, "(", nil, 1},
		{token.RIGHT_PAREN, ")", nil, 1},
		{token.LEFT_BRACE, "{", nil, 1},
		{token.RIGHT_BRACE, "}", nil, 1},
		{token.SEMICOLON, ";", nil, 1},
		{token.COMMA, ",", nil, 1},
		{token.PLUS, "+", nil, 1},
		{token.MINUS, "-", nil, 1},
		{token.STAR, "*", nil, 1},
		{token.BANG_EQUAL, "!=", nil, 1},
		{token.EQUAL_EQUAL, "==", nil, 1},
		{token.LESS_EQUAL, "<=", nil, 1},
		{token.GREATER_EQUAL, ">=", nil, 1},
		{token.BANG_EQUAL, "!=", nil, 1},
		{token.LESS, "<", nil, 1},
		{token.GREATER, ">", nil, 1},
		{token.SLASH, "/", nil, 1},
		{token.DOT, ".", nil, 1},
		{token.EOF, "", nil, 1},
	}

	runScanner(t, source, expected)
}

func TestStrings(t *testing.T) {
	source := `""
"string"`

	expected := []token.Token{
		{token.STRING, "\"\"", "", 1},
		{token.STRING, "\"string\"", "string", 2},
		{token.EOF, "", nil, 2},
	}

	runScanner(t, source, expected)
}

func TestWhitespace(t *testing.T) {
	source := `space    tabs				newlines




end`

	expected := []token.Token{
		{token.IDENTIFIER, "space", nil, 1},
		{token.IDENTIFIER, "tabs", nil, 1},
		{token.IDENTIFIER, "newlines", nil, 1},
		{token.IDENTIFIER, "end", nil, 6},
		{token.EOF, "", nil, 6},
	}

	runScanner(t, source, expected)
}
