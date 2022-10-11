package dsl

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	Invalid rune = -(iota + 1)
	Literal
	Command
	Reference
	Skip
	Date
	Datetime
	Number
	Boolean
	Comment
	Equal
	Colon
	Lparen
	Rparen
	Comma
	Blank
	EOL
	EOF
)

type Token struct {
	Literal string
	Type    rune
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	default:
		prefix = "unknown"
	case Invalid:
		prefix = "invalid"
	case Literal:
		prefix = "literal"
	case Command:
		prefix = "command"
	case Reference:
		prefix = "reference"
	case Skip:
		return "<skip>"
	case Date:
		prefix = "date"
	case Datetime:
		prefix = "datetime"
	case Number:
		prefix = "number"
	case Boolean:
		prefix = "boolean"
	case Comment:
		prefix = "comment"
	case Equal:
		return "<equal>"
	case Colon:
		return "<colon>"
	case Lparen:
		return "<lparen>"
	case Rparen:
		return "<rparen>"
	case Comma:
		return "<comma>"
	case Blank:
		return "<blank>"
	case EOL:
		return "<eol>"
	case EOF:
		return "<eof>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

type Scanner struct {
	input []byte

	curr int
	next int
	char rune

	keepBlank bool
}

func Scan(r io.Reader) *Scanner {
	in, _ := io.ReadAll(r)
	scan := Scanner{
		input: bytes.ReplaceAll(in, []byte{cr, nl}, []byte{nl}),
	}
	return &scan
}

func (s *Scanner) Scan() Token {
	s.read()

	var tok Token
	if s.done() {
		tok.Type = EOF
		return tok
	}
	switch {
	case isComment(s.char):
		s.scanComment(&tok)
	case isReference(s.char):
		s.scanReference(&tok)
	case isLetter(s.char):
		s.scanLiteral(&tok)
	case isQuote(s.char):
		s.scanQuote(&tok)
	case isDigit(s.char) || s.char == dash:
		s.scanNumber(&tok)
	case isBlank(s.char):
		s.scanBlank(&tok)
	case isNL(s.char):
		s.scanNewline(&tok)
	case isPunct(s.char):
		s.scanPunct(&tok)
	default:
		tok.Type = Invalid
	}
	return tok
}

func (s *Scanner) scanPunct(tok *Token) {
	switch s.char {
	case equal:
		tok.Type = Equal
	case colon:
		tok.Type = Colon
	case comma:
		tok.Type = Comma
	case lparen:
		tok.Type = Lparen
	case rparen:
		tok.Type = Rparen
	default:
	}
}

func (s *Scanner) scanBlank(tok *Token) {
	tok.Type = Blank
	s.skipBlank()
}

func (s *Scanner) scanNewline(tok *Token) {
	s.read()
	tok.Type = EOL
	if isBlank(s.char) {
		s.skipBlank()
		tok.Type = Blank
	} else {
		s.skipNL()
	}
}

func (s *Scanner) scanComment(tok *Token) {
	s.read()
	s.skipBlank()
	s.read()
	pos := s.curr
	for !isNL(s.char) {
		s.read()
	}
	tok.Type = Comment
	tok.Literal = string(s.input[pos:s.curr])
	s.skipNL()
}

func (s *Scanner) scanLiteral(tok *Token) {
	defer s.unread()

	pos := s.curr
	for isLetter(s.char) {
		s.read()
	}
	tok.Type = Literal
	tok.Literal = string(s.input[pos:s.curr])
	if tok.Literal == "true" || tok.Literal == "false" {
		tok.Type = Boolean
	}
}

func (s *Scanner) scanQuote(tok *Token) {
	quote := s.char
	s.read()
	pos := s.curr
	for s.char != quote {
		s.read()
	}
	tok.Type = Literal
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanCommand(tok *Token) {
	s.read()
	pos := s.curr
	for s.char != rparen {
		s.read()
	}
	tok.Type = Command
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanReference(tok *Token) {
	s.read()
	if s.char == lparen {
		s.scanCommand(tok)
		return
	}
	defer s.unread()
	pos := s.curr
	for isLetter(s.char) {
		s.read()
	}
	tok.Type = Reference
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanNumber(tok *Token) {
	defer s.unread()

	pos := s.curr
	if s.char == dash {
		s.read()
		if isBlank(s.char) {
			tok.Type = Skip
			return
		}
	}
	for isDigit(s.char) {
		s.read()
	}
	switch s.char {
	case dot:
		s.read()
		for isDigit(s.char) {
			s.read()
		}
	case dash:
		s.scanDate(pos, tok)
		return
	default:
	}
	tok.Type = Number
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanDate(pos int, tok *Token) {
	s.read()
	for isDigit(s.char) {
		s.read()
	}
	if s.char != dash {
		tok.Type = Invalid
		return
	}
	s.read()
	for isDigit(s.char) {
		s.read()
	}
	tok.Type = Date
	tok.Literal = string(s.input[pos:s.curr])
	if s.char == 'T' {
		s.scanTime(pos, tok)
	}
}

func (s *Scanner) scanTime(pos int, tok *Token) {
	s.read()
	for isDigit(s.char) {
		s.read()
	}
	if s.char != colon {
		tok.Type = Invalid
		return
	}
	s.read()
	for isDigit(s.char) {
		s.read()
	}
	if s.char != colon {
		tok.Type = Invalid
		return
	}
	s.read()
	for isDigit(s.char) {
		s.read()
	}
	tok.Type = Datetime
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) skipBlank() {
	s.skip(isBlank)
}

func (s *Scanner) skipNL() {
	s.skip(isNL)
}

func (s *Scanner) skip(accept func(rune) bool) {
	defer s.unread()
	for accept(s.char) {
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
	r, size := utf8.DecodeRune(s.input[s.next:])
	s.curr = s.next
	s.next += size
	s.char = r
}

func (s *Scanner) unread() {
	s.next = s.curr
	s.curr = s.curr - utf8.RuneLen(s.char)
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
	equal           = '='
	lparen          = '('
	rparen          = ')'
	comma           = ','
	hash            = '#'
	dollar          = '$'
	dot             = '.'
	dash            = '-'
	squote          = '\''
	dquote          = '"'
	underscore      = '_'
)

func isPunct(r rune) bool {
	return r == equal || r == colon || r == lparen || r == rparen || r == comma
}

func isComment(r rune) bool {
	return r == hash
}

func isLetter(r rune) bool {
	return isLower(r) || isUpper(r) || r == dash || r == underscore
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

func isReference(r rune) bool {
	return r == dollar
}

func isBlank(r rune) bool {
	return r == space || r == tab
}

func isNL(r rune) bool {
	return r == nl
}
