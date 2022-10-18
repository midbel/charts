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
	kwLimit   = "limit"
	kwInclude = "include"
	kwDefine  = "define"
	// kwTo      = "to"
	// kwAs      = "as"
)

func isKeyword(str string) bool {
	switch str {
	default:
		return false
	case kwSet:
	case kwLoad:
	case kwUsing:
	case kwRender:
	case kwWith:
	case kwLimit:
	case kwInclude:
	case kwDefine:
	}
	return true
}

const (
	Invalid rune = -(iota + 1)
	Keyword
	Literal
	Ident
	Variable
	Number
	Command
	Comment
	Comma
	Lparen
	Rparen
	Sum
	Range
	RangeSum
	Add
	Sub
	Mul
	Pow
	Div
	Mod
	Assign
	EOL
	EOF
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Literal string
	Type    rune
	Position
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
	case Number:
		prefix = "number"
	case Comment:
		prefix = "comment"
	case Keyword:
		prefix = "keyword"
	case Variable:
		prefix = "variable"
	case Ident:
		prefix = "identifier"
	case Command:
		prefix = "command"
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
	case Sum:
		return "<sum>"
	case Range:
		return "<range>"
	case RangeSum:
		return "<range-sum>"
	case Add:
		return "<add>"
	case Sub:
		return "<subtract>"
	case Mul:
		return "<multiply>"
	case Div:
		return "<divide>"
	case Mod:
		return "<modulo>"
	case Pow:
		return "<power>"
	case Assign:
		return "<assign>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

type Lexer struct {
	input []byte

	curr int
	next int
	char rune
}

func Lex(r io.Reader) *Lexer {
	in, _ := io.ReadAll(r)
	x := Lexer{
		input: bytes.ReplaceAll(in, []byte{cr, nl}, []byte{nl}),
	}
	return &x
}

func (x *Lexer) Lex() Token {
	x.read()
	if isBlank(x.char) {
		x.skipBlank()
		x.read()
	}
	var tok Token
	if x.done() {
		tok.Type = EOF
		return tok
	}
	switch {
	case isDigit(x.char):
		x.lexNumber(&tok)
	case isOperator(x.char):
		x.lexOperator(&tok)
	case isDollar(x.char):
		x.lexVariable(&tok)
	case isChar(x.char):
		x.lexIdent(&tok)
	case isQuote(x.char):
		x.lexLiteral(&tok)
	case isNL(x.char):
		x.skipNL()
		tok.Type = EOL
	default:
		tok.Type = Invalid
	}
	return tok
}

func (x *Lexer) lexLiteral(tok *Token) {
	quote := x.char
	x.read()
	pos := x.curr
	for x.char != quote && !x.done() {
		x.read()
	}
	tok.Type = Literal
	tok.Literal = string(x.input[pos:x.curr])
}

func (x *Lexer) lexIdent(tok *Token) {
	defer x.unread()
	pos := x.curr
	for isLetter(x.char) {
		x.read()
	}
	tok.Type = Ident
	tok.Literal = string(x.input[pos:x.curr])
}

func (x *Lexer) lexNumber(tok *Token) {
	defer x.unread()

	pos := x.curr
	for isDigit(x.char) {
		x.read()
	}
	if x.char == dot {
		x.read()
		for isDigit(x.char) {
			x.read()
		}
	}
	tok.Type = Number
	tok.Literal = string(x.input[pos:x.curr])
}

func (x *Lexer) lexVariable(tok *Token) {
	defer x.unread()

	x.read()
	if !isChar(x.char) {
		tok.Type = Invalid
		return
	}
	pos := x.curr
	for isChar(x.char) || isDigit(x.char) {
		x.read()
	}
	tok.Type = Variable
	tok.Literal = string(x.input[pos:x.curr])
}

func (x *Lexer) lexOperator(tok *Token) {
	switch x.char {
	case lparen:
		tok.Type = Lparen
	case rparen:
		tok.Type = Rparen
	case plus:
		tok.Type = Add
	case minus:
		tok.Type = Sub
	case star:
		tok.Type = Mul
		if x.peek() == star {
			tok.Type = Pow
			x.read()
		}
	case slash:
		tok.Type = Div
	case percent:
		tok.Type = Mod
	case equal:
		tok.Type = Assign
	case semicolon:
		tok.Type = EOL
	default:
		tok.Type = Invalid
	}
}

func (x *Lexer) done() bool {
	return x.char == utf8.RuneError
}

func (x *Lexer) peek() rune {
	r, _ := utf8.DecodeRune(x.input[x.next:])
	return r
}

func (x *Lexer) read() {
	if x.curr >= len(x.input) || x.char == utf8.RuneError {
		return
	}
	r, size := utf8.DecodeRune(x.input[x.next:])
	x.curr = x.next
	x.next += size
	x.char = r
}

func (x *Lexer) unread() {
	var size int
	x.char, size = utf8.DecodeRune(x.input[x.curr:])
	x.next = x.curr
	x.curr -= size
}

func (x *Lexer) skipBlank() {
	x.skip(isBlank)
}

func (x *Lexer) skipNL() {
	x.skip(isNL)
}

func (x *Lexer) skip(accept func(rune) bool) {
	defer x.unread()
	for accept(x.char) && !x.done() {
		x.read()
	}
}

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
	default:
		s.scanLiteral(&tok)
	}
	return tok
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
	minus           = '-'
	slash           = '/'
	star            = '*'
	percent         = '%'
	semicolon       = ';'
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

func isDollar(r rune) bool {
	return r == dollar
}

func isPunct(r rune) bool {
	return r == comma || r == lparen || r == rparen || r == colon || r == plus
}

func isOperator(r rune) bool {
	switch r {
	case plus, minus, star, percent, slash, semicolon, lparen, rparen, equal:
		return true
	default:
		return false
	}
}

func isChar(r rune) bool {
	return isLower(r) || isUpper(r)
}

func isLetter(r rune) bool {
	return isChar(r) || r == dash || r == underscore
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
