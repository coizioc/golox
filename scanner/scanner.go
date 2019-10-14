package scanner

import (
	"fmt"
	"golox/loxerror"
	"golox/token"
	"strconv"
	"strings"
)

type Scanner struct {
	Source  string
	Tokens  []token.Token
	Start   int
	Current int
	Line    int
}

func New(source string) *Scanner {
	return &Scanner{source, []token.Token{}, 0, 0, 1}
}

func (sc *Scanner) ScanTokens() []token.Token {
	for !sc.isAtEnd() {
		sc.Start = sc.Current
		sc.scanToken()
	}

	sc.Tokens = append(sc.Tokens, token.Token{token.EOF, "", nil, sc.Line})
	return sc.Tokens
}

func (sc *Scanner) String() string {
	sb := strings.Builder{}
	for i, tok := range sc.Tokens {
		sb.WriteString(fmt.Sprintf("%3d: %s\n", i, tok.String()))
	}
	return sb.String()
}

func (sc *Scanner) isAtEnd() bool {
	return sc.Current >= len(sc.Source)
}

func (sc *Scanner) advance() byte {
	sc.Current++
	return sc.Source[sc.Current-1]
}

func (sc *Scanner) match(expected byte) bool {
	if sc.isAtEnd() {
		return false
	}
	if sc.Source[sc.Current] != expected {
		return false
	}
	sc.Current++
	return true
}

func (sc *Scanner) peek() byte {
	if sc.isAtEnd() {
		return 0
	}
	return sc.Source[sc.Current]
}

func (sc *Scanner) peekNext() byte {
	if sc.Current+1 >= len(sc.Source) {
		return 0
	}
	return sc.Source[sc.Current+1]
}

func (sc *Scanner) isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c == '_'
}

func (sc *Scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (sc *Scanner) isAlphanumeric(c byte) bool {
	return sc.isAlpha(c) || sc.isDigit(c)
}

func (sc *Scanner) handleIdentifier() {
	for sc.isAlphanumeric(sc.peek()) {
		sc.advance()
	}

	// Handle keywords
	text := sc.Source[sc.Start:sc.Current]
	tokenType, ok := keywords[text]
	if ok {
		sc.addToken(tokenType, nil)
	} else {
		sc.addToken(token.IDENTIFIER, nil)
	}
}

func (sc *Scanner) handleNumber() {
	for sc.isDigit(sc.peek()) {
		sc.advance()
	}

	if sc.peek() == '.' && sc.isDigit(sc.peekNext()) {
		// Consume the '.'
		sc.advance()
		for sc.isDigit(sc.peek()) {
			sc.advance()
		}
	}

	number, _ := strconv.ParseFloat(sc.Source[sc.Start:sc.Current], 64)

	sc.addToken(token.NUMBER, number)
}

func (sc *Scanner) handleString() {
	for sc.peek() != '"' && !sc.isAtEnd() {
		if sc.peek() == '\n' {
			sc.Line++
		}
		sc.advance()
	}

	if sc.isAtEnd() {
		loxerror.Error(sc.Line, "Unterminated string.")
		return
	}

	// For the closing ".
	sc.advance()

	value := sc.Source[sc.Start+1 : sc.Current-1]
	sc.addToken(token.STRING, value)
}

func (sc *Scanner) addToken(tokenType token.TokenType, literal interface{}) {
	text := sc.Source[sc.Start:sc.Current]
	sc.Tokens = append(sc.Tokens, token.Token{tokenType, text, literal, sc.Line})
}

func (sc *Scanner) scanToken() {
	c := sc.advance()
	switch c {
	case '(':
		sc.addToken(token.LEFT_PAREN, nil)
	case ')':
		sc.addToken(token.RIGHT_PAREN, nil)
	case '{':
		sc.addToken(token.LEFT_BRACE, nil)
	case '}':
		sc.addToken(token.RIGHT_BRACE, nil)
	case ',':
		sc.addToken(token.COMMA, nil)
	case '.':
		sc.addToken(token.DOT, nil)
	case '-':
		sc.addToken(token.MINUS, nil)
	case '+':
		sc.addToken(token.PLUS, nil)
	case ';':
		sc.addToken(token.SEMICOLON, nil)
	case '*':
		sc.addToken(token.STAR, nil)
	case '!':
		if sc.match('=') {
			sc.addToken(token.BANG_EQUAL, nil)
		} else {
			sc.addToken(token.BANG, nil)
		}
	case '=':
		if sc.match('=') {
			sc.addToken(token.EQUAL_EQUAL, nil)
		} else {
			sc.addToken(token.EQUAL, nil)
		}
	case '<':
		if sc.match('=') {
			sc.addToken(token.LESS_EQUAL, nil)
		} else {
			sc.addToken(token.LESS, nil)
		}
	case '>':
		if sc.match('=') {
			sc.addToken(token.GREATER_EQUAL, nil)
		} else {
			sc.addToken(token.GREATER, nil)
		}
	case '/':
		// Handle line comments
		if sc.match('/') {
			for sc.peek() != '\n' && !sc.isAtEnd() {
				sc.advance()
			}
			// Handle block comments
		} else if sc.match('*') {
			for sc.peek() != '*' && sc.peekNext() != '/' && !sc.isAtEnd() {
				if sc.peek() == '\n' {
					sc.Line++
				}
				sc.advance()
			}
			if sc.isAtEnd() {
				loxerror.Error(sc.Line, "Unterminated block comment.")
			}
		} else {
			sc.addToken(token.SLASH, nil)
		}
	case ' ':
	case '\r':
	case '\t':

	case '\n':
		sc.Line++
	case '"':
		sc.handleString()
	default:
		if sc.isDigit(c) {
			sc.handleNumber()
		} else if sc.isAlpha(c) {
			sc.handleIdentifier()
		} else {
			loxerror.Error(sc.Line, "Unexpected character.")
		}
	}
}
