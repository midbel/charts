package dsl

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	kwSet     = "set"
	kwLoad    = "load"
	kwUsing   = "using"
	kwRender  = "render"
	kwWith    = "with"
	kwInclude = "include"
)

const (
	Invalid rune = -(iota + 1)
	Keyword
	Literal
	Comment
	Comma
	Lparen
	Rparen
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
	case Comment:
		prefix = "comment"
	case Keyword:
		prefix = "keyword"
	case Comma:
		return "<comma>"
	case EOL:
		return "<eol>"
	case EOF:
		return "<eof>"
	case Lparen:
		return "<lparen>"
	case Rparen:
		return "<rparen>"
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
	if isBlank(s.char) {
		s.skipBlank()
	}

	var tok Token
	if s.done() {
		tok.Type = EOF
		return tok
	}
	switch {
	case isComment(s.char):
		s.scanComment(&tok)
	case isQuote(s.char):
		s.scanQuote(&tok)
	case isNL(s.char):
		s.scanNewline(&tok)
	case isPunct(s.char):
		s.scanPunct(&tok)
	default:
		s.scanLiteral(&tok)
	}
	return tok
}

func (s *Scanner) scanPunct(tok *Token) {
	switch s.char {
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
	s.skipNL()
	tok.Type = EOL
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
	pos := s.curr
	for !isBlank(s.char) && !isPunct(s.char) && !isNL(s.char) && !s.done() {
		s.read()
	}
	tok.Type = Literal
	tok.Literal = string(s.input[pos:s.curr])
	if !isBlank(s.char) {
		s.unread()
	}
	switch tok.Literal {
	default:
	case kwSet, kwLoad, kwRender, kwUsing, kwWith, kwInclude:
		tok.Type = Keyword
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
	return r == comma || r == lparen || r == rparen
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

func isBlank(r rune) bool {
	return r == space || r == tab
}

func isNL(r rune) bool {
	return r == nl
}
