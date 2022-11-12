package decode

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"
)

type Scanner struct {
	input []byte

	curr int
	next int
	char rune

	Position
	seen int
}

func Scan(r io.Reader) *Scanner {
	in, _ := io.ReadAll(r)
	sc := Scanner{
		input: bytes.ReplaceAll(in, []byte{cr, nl}, []byte{nl}),
	}
	sc.Line++
	return &sc
}

func (s *Scanner) Scan() Token {
	s.read()
	if isBlank(s.char) {
		s.skipBlank()
		s.read()
	}

	var tok Token
	tok.Position = s.Position
	if s.done() {
		tok.Type = EOF
		return tok
	}
	switch {
	case isDollar(s.char):
		s.scanCommand(&tok)
	case isQuote(s.char):
		s.scanQuote(&tok)
	case isNL(s.char):
		s.scanNewline(&tok)
	case isPunct(s.char):
		s.scanPunct(&tok)
	case isScript(s.char):
		s.scanScript(&tok)
	default:
		s.scanLiteral(&tok)
	}
	return tok
}

func (s *Scanner) scanScript(tok *Token) {
	var consume func(int) bool
	consume = func(level int) bool {
		for s.char != rcurly && !s.done() {
			s.read()
			if s.char == lcurly {
				s.read()
				if ok := consume(level + 1); !ok {
					return ok
				}
			}
		}
		if s.char != rcurly {
			return false
		}
		if level > 0 {
			s.read()
		}
		return true
	}
	s.read()
	var (
		pos   = s.curr
		valid = consume(0)
	)

	str := string(s.input[pos:s.curr])
	tok.Type = Expr
	tok.Literal = strings.TrimSpace(str)
	if s.char != rcurly || !valid {
		tok.Type = Invalid
	}
}

func (s *Scanner) scanVariable(tok *Token) {
	pos := s.curr
	for !isBlank(s.char) && !isPunct(s.char) && !isNL(s.char) && !s.done() {
		s.read()
	}
	tok.Type = Variable
	tok.Literal = string(s.input[pos:s.curr])
	if !isBlank(s.char) {
		s.unread()
	}
}

func (s *Scanner) scanCommand(tok *Token) {
	s.read()
	if s.char != lparen {
		s.scanVariable(tok)
		return
	}
	s.read()
	pos := s.curr
	for s.char != rparen && !s.done() {
		s.read()
	}
	tok.Type = Command
	tok.Literal = string(s.input[pos:s.curr])
	if s.char != rparen {
		tok.Type = Invalid
	}
}

func (s *Scanner) scanLiteral(tok *Token) {
	pos := s.curr
	for !isBlank(s.char) && !isPunct(s.char) && !isNL(s.char) && !s.done() {
		s.read()
	}
	tok.Type = Literal
	tok.Literal = string(s.input[pos:s.curr])
	if !isBlank(s.char) {
		s.unread()
	}
	if isKeyword(tok.Literal) {
		tok.Type = Keyword
	}
}

func (s *Scanner) scanPunct(tok *Token) {
	switch s.char {
	case colon:
		tok.Type = Range
		if s.peek() == plus {
			tok.Type = RangeSum
			s.read()
		}
	case plus:
		tok.Type = Sum
	case comma:
		tok.Type = Comma
		s.read()
		s.skipBlank()
	case lparen:
		tok.Type = Lparen
	case rparen:
		tok.Type = Rparen
	default:
		tok.Type = Invalid
	}
}

func (s *Scanner) scanNewline(tok *Token) {
	if isNL(s.peek()) {
		s.skipNL()
	}
	tok.Type = EOL
}

func (s *Scanner) scanQuote(tok *Token) {
	quote := s.char
	s.read()
	pos := s.curr
	for s.char != quote && !s.done() {
		s.read()
	}
	tok.Type = Literal
	tok.Literal = string(s.input[pos:s.curr])
	if s.char != quote {
		tok.Type = Invalid
	}
}

func (s *Scanner) skipBlank() {
	s.skip(isBlank)
}

func (s *Scanner) skipNL() {
	s.skip(isNL)
}

func (s *Scanner) skip(accept func(rune) bool) {
	defer s.unread()
	for accept(s.char) && !s.done() {
		s.read()
	}
}

func (s *Scanner) done() bool {
	return s.char == utf8.RuneError
}

func (s *Scanner) read() {
	if s.curr >= len(s.input) || s.char == utf8.RuneError {
		return
	}
	if s.char == nl {
		s.seen = s.Column
		s.Line++
		s.Column = 0
	}
	s.Column++

	r, size := utf8.DecodeRune(s.input[s.next:])
	s.curr = s.next
	s.next += size
	s.char = r
}

func (s *Scanner) unread() {
	var size int
	if s.char == nl {
		s.Line--
		s.Column = s.seen
	}
	s.Column--
	s.char, size = utf8.DecodeRune(s.input[s.curr:])
	s.next = s.curr
	s.curr -= size
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

const (
	space      rune = ' '
	tab             = '\t'
	cr              = '\r'
	nl              = '\n'
	colon           = ':'
	plus            = '+'
	percent         = '%'
	lparen          = '('
	rparen          = ')'
	lcurly          = '{'
	rcurly          = '}'
	comma           = ','
	hash            = '#'
	dollar          = '$'
	dot             = '.'
	minus           = '-'
	squote          = '\''
	dquote          = '"'
	underscore      = '_'
)

func isDollar(r rune) bool {
	return r == dollar
}

func isPunct(r rune) bool {
	return r == comma || r == lparen || r == rparen || r == colon || r == plus
}

func isScript(r rune) bool {
	return r == lcurly
}

func isChar(r rune) bool {
	return isLower(r) || isUpper(r)
}

func isLetter(r rune) bool {
	return isChar(r) || r == minus || r == underscore
}

func isAlpha(r rune) bool {
	return isLetter(r) || isDigit(r)
}

func isQuote(r rune) bool {
	return r == squote || r == dquote
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isBlank(r rune) bool {
	return r == space || r == tab
}

func isNL(r rune) bool {
	return r == nl
}
