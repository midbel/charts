package decode

import (
	"fmt"
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
	kwDeclare = "declare"
	kwAt      = "at"
	kwUse     = "use"
	kwTo      = "to"
	kwAs      = "as"
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
	case kwDeclare:
	case kwAt:
	case kwUse:
	case kwAs:
	case kwTo:
	}
	return true
}

const (
	Invalid rune = -(iota + 1)
	Keyword
	Literal
	Variable
	Command
	Data
	Comment
	Comma
	Lparen
	Rparen
	Lcurly
	Rcurly
	Expr
	Sum
	Range
	RangeSum
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
	case Expr:
		prefix = "expression"
	case Comment:
		prefix = "comment"
	case Keyword:
		prefix = "keyword"
	case Variable:
		prefix = "variable"
	case Command:
		prefix = "command"
	case Data:
		prefix = "data"
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
	case Lcurly:
		return "<lcurly>"
	case Rcurly:
		return "<rcurly>"
	case Sum:
		return "<sum>"
	case Range:
		return "<range>"
	case RangeSum:
		return "<range-sum>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}
