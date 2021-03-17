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

type Scanner struct {
	source  string
	tokens  []*tokenObj
	start   int // start of the lexeme
	current int // pointer of scanner
	line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: source,
		tokens: make([]*tokenObj, 0),
		line:   1,
	}
}

func (s *Scanner) scan() []*tokenObj {
	for !s.atEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, &tokenObj{typ: EOF})
	return s.tokens
}

func (s *Scanner) scanToken() {
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
		} else if s.match('*') {
			s.fullComment()
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

func (s *Scanner) atEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) advance() byte {
	i := s.current
	s.current++
	return s.source[i]
}

func (s *Scanner) match(ch byte) bool {
	if s.peek() != ch {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) peek() byte {
	if s.atEnd() {
		return byte(0)
	}
	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return byte(0)
	}
	return s.source[s.current+1]
}

func (s *Scanner) addToken(t token) {
	s.addLiteral(t, nil)
}

func (s *Scanner) addLiteral(t token, literal interface{}) {
	lex := s.source[s.start:s.current]
	s.tokens = append(s.tokens, &tokenObj{
		typ:     t,
		lexeme:  lex,
		literal: literal,
		line:    s.line,
	})
}

func (s *Scanner) readString() {
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

func (s *Scanner) number() {
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

func (s *Scanner) identifier() {
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

func (s *Scanner) fullComment() {
	for !(s.peek() == '*' && s.peekNext() == '/') && !s.atEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	if s.atEnd() {
		report(s.line, "unterminated /**/ comment")
		return
	}
	s.advance() // skip *
	s.advance() // skip /
}
