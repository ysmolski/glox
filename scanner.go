package main

import "strconv"

var keywords = map[string]token{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"for":    For,
	"fun":    Fun,
	"if":     If,
	"nil":    Nil,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
}

type scanner struct {
	source  string
	tokens  []*tokenObj
	start   int // start of the lexeme
	current int // pointer of scanner
	line    int
}

func NewScanner(source string) *scanner {
	return &scanner{
		source: source,
		tokens: make([]*tokenObj, 0),
		line:   1,
	}
}

func (s *scanner) scan() []*tokenObj {
	for !s.atEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, &tokenObj{typ: EOF})
	return s.tokens
}

func (s *scanner) scanToken() {
	ch := s.advance()
	switch ch {
	case '(':
		s.addToken(LeftParen)
	case ')':
		s.addToken(RightParen)
	case '{':
		s.addToken(LeftBrace)
	case '}':
		s.addToken(RightBrace)
	case ',':
		s.addToken(Comma)
	case '.':
		s.addToken(Dot)
	case '-':
		s.addToken(Minus)
	case '+':
		s.addToken(Plus)
	case ';':
		s.addToken(Semicolon)
	case '*':
		s.addToken(Star)
	case '!':
		if s.match('=') {
			s.addToken(BangEqual)
		} else {
			s.addToken(Bang)
		}
	case '=':
		if s.match('=') {
			s.addToken(EqualEqual)
		} else {
			s.addToken(Equal)
		}
	case '<':
		if s.match('=') {
			s.addToken(LessEqual)
		} else {
			s.addToken(Less)
		}
	case '>':
		if s.match('=') {
			s.addToken(GreaterEqual)
		} else {
			s.addToken(Greater)
		}
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.atEnd() {
				s.advance()
			}
		} else {
			s.addToken(Slash)
		}
	case ' ', '\r', '\t':
		break
	case '\n':
		s.line++
	case '"':
		s.readString()
	default:
		if isDigit(ch) {
			s.number()
		} else if isAlpha(ch) {
			s.identifier()
		} else {
			report(s.line, "unexpected character")
		}
	}
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func isAlpha(b byte) bool {
	return 'a' <= b && b <= 'z' ||
		'A' <= b && b <= 'Z' ||
		b == '_'
}

func isAlphaNum(b byte) bool {
	return isDigit(b) || isAlpha(b)
}

func (s *scanner) atEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) advance() byte {
	i := s.current
	s.current++
	return s.source[i]
}

func (s *scanner) match(ch byte) bool {
	if s.peek() != ch {
		return false
	}
	s.current++
	return true
}

func (s *scanner) peek() byte {
	if s.atEnd() {
		return byte(0)
	}
	return s.source[s.current]
}

func (s *scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return byte(0)
	}
	return s.source[s.current+1]
}

func (s *scanner) addToken(t token) {
	s.addLiteral(t, nil)
}

func (s *scanner) addLiteral(t token, literal interface{}) {
	lex := s.source[s.start:s.current]
	s.tokens = append(s.tokens, &tokenObj{
		typ:     t,
		lexeme:  lex,
		literal: literal,
		line:    s.line,
	})
}

func (s *scanner) readString() {
	for s.peek() != '"' && !s.atEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	if s.atEnd() {
		report(s.line, "unterminated string")
		return
	}
	s.advance() // skip closing "
	value := s.source[s.start+1 : s.current-1]
	s.addLiteral(String, value)
}

func (s *scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}
	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance() // eat .
		for isDigit(s.peek()) {
			s.advance()
		}
	}
	val, err := strconv.ParseFloat(s.source[s.start:s.current], 64)
	if err != nil {
		report(s.line, "cannot parse float number")
	}
	s.addLiteral(Number, val)
}

func (s *scanner) identifier() {
	for isAlphaNum(s.peek()) {
		s.advance()
	}
	text := s.source[s.start:s.current]
	var t token
	if typ, ok := keywords[text]; ok {
		t = typ
	} else {
		t = Identifier
	}
	s.addToken(t)
}
