package dsl

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/slices"
)

func Eval(r io.Reader) (interface{}, error) {
	expr, err := Parse(r)
	if err != nil {
		return nil, err
	}
	var (
		env   = emptyEnv[any]()
		visit = []visitFunc{replaceValue}
	)
	if expr, err = traverse(expr, env, visit); err != nil {
		return nil, err
	}
	return eval(expr, env)
}

type visitFunc func(Expression, *environ[any]) (Expression, error)

func traverse(expr Expression, env *environ[any], visit []visitFunc) (Expression, error) {
	var err error
	for _, v := range visit {
		expr, err = v(expr, env)
		if err != nil {
			break
		}
	}
	return expr, err
}

func replaceExprList(list []Expression, env *environ[any]) ([]Expression, error) {
	var err error
	for i := range list {
		list[i], err = replaceValue(list[i], env)
		if err != nil {
			return nil, err
		}
	}
	return list, err
}

func replaceValue(expr Expression, env *environ[any]) (Expression, error) {
	var err error
	switch e := expr.(type) {
	case script:
		e.list, err = replaceExprList(e.list, env)
		return e, err
	case call:
		e.args, err = replaceExprList(e.args, env)
		return e, err
	case unary:
		e.right, err = replaceValue(e.right, env)
		if err != nil {
			return nil, err
		}
		if e.right.isPrimitive() {
			res, err := evalUnary(e, env)
			if err != nil {
				return nil, err
			}
			return createPrimitive(res)
		}
		return e, nil
	case binary:
		e.left, err = replaceValue(e.left, env)
		if err != nil {
			return nil, err
		}
		e.right, err = replaceValue(e.right, env)
		if err != nil {
			return nil, err
		}
		if e.left.isPrimitive() && e.right.isPrimitive() {
			res, err := evalBinary(e, env)
			if err != nil {
				return nil, err
			}
			return createPrimitive(res)
		}
		return e, nil
	case assign:
		e.right, err = replaceValue(e.right, env)
		return e, err
	case test:
		e.cdt, err = replaceValue(e.cdt, env)
		if err != nil {
			return nil, err
		}
		e.csq, err = replaceValue(e.csq, env)
		if err != nil {
			return nil, err
		}
		if e.alt == nil {
			return e, nil
		}
		e.alt, err = replaceValue(e.alt, env)
		return e, err
	case while:
		e.cdt, err = replaceValue(e.cdt, env)
		if err != nil {
			return nil, err
		}
		e.body, err = replaceValue(e.body, env)
		return e, err
	default:
		return expr, nil
	}
}

func eval(expr Expression, env *environ[any]) (interface{}, error) {
	var (
		res interface{}
		err error
	)
	switch e := expr.(type) {
	case script:
		for _, e := range e.list {
			res, err = eval(e, env)
			if err != nil {
				break
			}
		}
	case call:
		res, err = evalCall(e, env)
	case literal:
		return e.str, nil
	case number:
		return e.value, nil
	case boolean:
		return e.value, nil
	case variable:
		return env.Resolve(e.ident)
	case unary:
		res, err = evalUnary(e, env)
	case binary:
		res, err = evalBinary(e, env)
	case assign:
		res, err = evalAssign(e, env)
	case test:
		res, err = evalTest(e, env)
	case while:
		res, err = evalWhile(e, env)
	}
	return res, err
}

func evalUnary(u unary, env *environ[any]) (interface{}, error) {
	res, err := eval(u.right, env)
	if err != nil {
		return nil, err
	}
	switch u.op {
	case Not:
		return !isTruthy(res), nil
	case Sub:
		f, ok := res.(float64)
		if !ok {
			return nil, fmt.Errorf("expected float")
		}
		return -f, nil
	default:
		return nil, fmt.Errorf("unsupported unary operator")
	}
}

func evalBinary(b binary, env *environ[any]) (interface{}, error) {
	left, err := eval(b.left, env)
	if err != nil {
		return nil, err
	}
	right, err := eval(b.right, env)
	if err != nil {
		return nil, err
	}
	switch b.op {
	default:
		return nil, fmt.Errorf("unsupported binary operator")
	case Add:
		return execAdd(left, right)
	case Sub:
		return execSub(left, right)
	case Mul:
		return execMul(left, right)
	case Div:
		return execDiv(left, right)
	case Pow:
		return execPow(left, right)
	case Mod:
		return execMod(left, right)
	case And:
		return execAnd(left, right)
	case Or:
		return execOr(left, right)
	case Eq:
		return execEqual(left, right, false)
	case Ne:
		return execEqual(left, right, true)
	case Lt:
		return execLesser(left, right, false)
	case Le:
		return execLesser(left, right, true)
	case Gt:
		return execGreater(left, right, false)
	case Ge:
		return execGreater(left, right, true)
	}
}

func evalTest(t test, env *environ[any]) (interface{}, error) {
	res, err := eval(t.cdt, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(res) {
		return eval(t.csq, env)
	}
	if t.alt == nil {
		return nil, nil
	}
	return eval(t.alt, env)
}

func evalWhile(w while, env *environ[any]) (interface{}, error) {
	var (
		res interface{}
		err error
	)
	for {
		res, err = eval(w.cdt, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(res) {
			break
		}
		res, err = eval(w.body, env)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func evalAssign(a assign, env *environ[any]) (interface{}, error) {
	res, err := eval(a.right, env)
	if err != nil {
		return nil, err
	}
	env.Define(a.ident, res)
	return nil, nil
}

func evalCall(c call, env *environ[any]) (interface{}, error) {
	var (
		args []interface{}
		res  interface{}
		err  error
	)
	for _, a := range c.args {
		res, err = eval(a, env)
		if err != nil {
			return nil, err
		}
		args = append(args, res)
	}
	switch c.ident {
	case "len":
		if len(args) != 1 {
			return nil, fmt.Errorf("len: no enough argument given")
		}
		str, ok := slices.Fst(args).(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type: string expected")
		}
		return float64(len(str)), nil
	case "lower":
		if len(args) < 1 {
			return nil, fmt.Errorf("printf: no enough argument given")
		}
		str, ok := slices.Fst(args).(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type: string expected")
		}
		return strings.ToLower(str), nil
	case "upper":
		if len(args) < 1 {
			return nil, fmt.Errorf("printf: no enough argument given")
		}
		str, ok := slices.Fst(args).(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type: string expected")
		}
		return strings.ToUpper(str), nil
	case "printf", "format":
		if len(args) < 1 {
			return nil, fmt.Errorf("printf: no enough argument given")
		}
		pattern, ok := slices.Fst(args).(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type: string expected")
		}
		res = fmt.Sprintf(pattern, slices.Rest(args)...)
	case "print":
		if len(args) < 1 {
			return nil, fmt.Errorf("printf: no enough argument given")
		}
		res = fmt.Sprint(args...)
	case "time":
		if len(args) != 0 {
			return nil, fmt.Errorf("time: too many arguments given")
		}
		t := time.Now().Unix()
		return float64(t), nil
	default:
		return nil, fmt.Errorf("%s: function undefined", c.ident)
	}
	return res, nil
}

func execLesser(left, right interface{}, eq bool) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		y, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isLesser(x, y, eq), nil
	case string:
		y, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isLesser(x, y, eq), nil
	default:
		return nil, fmt.Errorf("type can not be compared")
	}
}

func execGreater(left, right interface{}, eq bool) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		y, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isGreater(x, y, eq), nil
	case string:
		y, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isGreater(x, y, eq), nil
	default:
		return nil, fmt.Errorf("type can not be compared")
	}
}

func execEqual(left, right interface{}, ne bool) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		y, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("incompatible type for equality")
		}
		return isEqual(x, y, ne), nil
	case string:
		y, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type for equality")
		}
		return isEqual(x, y, ne), nil
	case bool:
		y, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("incompatible type for equality")
		}
		return isEqual(x, y, ne), nil
	default:
		return nil, fmt.Errorf("type can not be compared")
	}
	return nil, nil
}

func isEqual[T float64 | string | bool](left, right T, ne bool) bool {
	ok := left == right
	if ne {
		ok = !ok
	}
	return ok
}

func isLesser[T float64 | string](left, right T, eq bool) bool {
	ok := left < right
	if !ok && eq {
		ok = left == right
	}
	return ok
}

func isGreater[T float64 | string](left, right T, eq bool) bool {
	ok := left > right
	if !ok && eq {
		ok = left == right
	}
	return ok
}

func execAnd(left, right interface{}) (interface{}, error) {
	return isTruthy(left) && isTruthy(right), nil
}
func execOr(left, right interface{}) (interface{}, error) {
	return isTruthy(left) || isTruthy(right), nil
}

func execAdd(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			return x + y, nil
		}
		if y, ok := right.(string); ok {
			return fmt.Sprintf("%f%s", x, y), nil
		}
		return nil, fmt.Errorf("incompatible type for addition")
	case string:
		if y, ok := right.(float64); ok {
			return fmt.Sprintf("%s%f", x, y), nil
		}
		if y, ok := right.(string); ok {
			return x + y, nil
		}
		return nil, fmt.Errorf("incompatible type for addition")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execSub(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			return x - y, nil
		}
		return nil, fmt.Errorf("incompatible type for subtraction")
	case string:
		if y, ok := right.(float64); ok {
			if y < 0 && int(math.Abs(y)) < len(x) {
				y = math.Abs(y)
				return x[int(y):], nil
			}
			if y > 0 && int(y) < len(x) {
				return x[:int(y)], nil
			}
		}
		return nil, fmt.Errorf("incompatible type for subtraction")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execMul(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			return x * y, nil
		}
		if y, ok := right.(string); ok {
			return strings.Repeat(y, int(x)), nil
		}
		return nil, fmt.Errorf("incompatible type for multiply")
	case string:
		if y, ok := right.(float64); ok {
			return strings.Repeat(x, int(y)), nil
		}
		return nil, fmt.Errorf("incompatible type for multiply")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execDiv(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			if y < 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return x / y, nil
		}
		return nil, fmt.Errorf("incompatible type for division")
	case string:
		if y, ok := right.(float64); ok && y > 0 {
			z := len(x) / int(y)
			return x[:z], nil
		}
		return nil, fmt.Errorf("incompatible type for division")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execMod(left, right interface{}) (interface{}, error) {
	x, ok1 := left.(float64)
	y, ok2 := right.(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("incompatible type for modulo")
	}
	if y == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return math.Mod(x, y), nil
}

func execPow(left, right interface{}) (interface{}, error) {
	x, ok1 := left.(float64)
	y, ok2 := right.(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("incompatible type for power")
	}
	return math.Pow(x, y), nil
}

func isTruthy(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		return x != ""
	default:
		return v != nil
	}
}

type Expression interface {
	isPrimitive() bool
}

type script struct {
	list []Expression
}

func (_ script) isPrimitive() bool {
	return false
}

type call struct {
	ident string
	args  []Expression
}

func (_ call) isPrimitive() bool {
	return false
}

type while struct {
	cdt  Expression
	body Expression
}

func (_ while) isPrimitive() bool {
	return false
}

type test struct {
	cdt Expression
	csq Expression
	alt Expression
}

func (_ test) isPrimitive() bool {
	return false
}

func createPrimitive(res interface{}) (Expression, error) {
	switch r := res.(type) {
	case float64:
		return createNumber(r), nil
	case bool:
		return createBoolean(r), nil
	case string:
		return createLiteral(r), nil
	default:
		return nil, fmt.Errorf("unexpected returned type from unary")
	}
}

type literal struct {
	str string
}

func createLiteral(str string) literal {
	return literal{
		str: str,
	}
}

func (_ literal) isPrimitive() bool {
	return true
}

type variable struct {
	ident string
}

func createVariable(ident string) variable {
	return variable{
		ident: ident,
	}
}

func (_ variable) isPrimitive() bool {
	return false
}

type boolean struct {
	value bool
}

func createBoolean(b bool) boolean {
	return boolean{
		value: b,
	}
}

func (_ boolean) isPrimitive() bool {
	return true
}

type number struct {
	value float64
}

func createNumber(f float64) number {
	return number{
		value: f,
	}
}

func (_ number) isPrimitive() bool {
	return true
}

type unary struct {
	op    rune
	right Expression
}

func (_ unary) isPrimitive() bool {
	return false
}

type binary struct {
	op    rune
	left  Expression
	right Expression
}

func (_ binary) isPrimitive() bool {
	return false
}

type assign struct {
	ident string
	right Expression
}

func (_ assign) isPrimitive() bool {
	return false
}

const (
	powLowest   = iota
	powAssign   // =
	powTernary  // ?:
	powRelation // &&, ||
	powEqual    // ==, !=
	powCompare  // <, <=, >, >=
	powAdd      // +, -
	powMul      // /, *, **, %
	powPrefix
	powCall // ()
)

type powerMap map[rune]int

func (p powerMap) Get(r rune) int {
	v, ok := p[r]
	if !ok {
		return powLowest
	}
	return v
}

var powers = powerMap{
	Add:       powAdd,
	Sub:       powAdd,
	Mul:       powMul,
	Div:       powMul,
	Mod:       powMul,
	Pow:       powMul,
	Assign:    powAssign,
	AddAssign: powAssign,
	SubAssign: powAssign,
	MulAssign: powAssign,
	DivAssign: powAssign,
	ModAssign: powAssign,
	Lparen:    powCall,
	Ternary:   powTernary,
	And:       powRelation,
	Or:        powRelation,
	Eq:        powEqual,
	Ne:        powEqual,
	Lt:        powCompare,
	Le:        powCompare,
	Gt:        powCompare,
	Ge:        powCompare,
}

type parser struct {
	lex  *Lexer
	curr Token
	peek Token

	prefix map[rune]func() (Expression, error)
	infix  map[rune]func(Expression) (Expression, error)
}

func Parse(r io.Reader) (Expression, error) {
	p := parser{
		lex: Lex(r),
	}
	p.prefix = map[rune]func() (Expression, error){
		Sub:     p.parsePrefix,
		Not:     p.parsePrefix,
		Number:  p.parsePrefix,
		Boolean: p.parsePrefix,
		Literal: p.parsePrefix,
		Ident:   p.parsePrefix,
		Lparen:  p.parseGroup,
		Keyword: p.parseKeyword,
	}
	p.infix = map[rune]func(Expression) (Expression, error){
		Add:       p.parseInfix,
		Sub:       p.parseInfix,
		Mul:       p.parseInfix,
		Div:       p.parseInfix,
		Mod:       p.parseInfix,
		Pow:       p.parseInfix,
		Assign:    p.parseAssign,
		AddAssign: p.parseAssign,
		SubAssign: p.parseAssign,
		DivAssign: p.parseAssign,
		MulAssign: p.parseAssign,
		ModAssign: p.parseAssign,
		Lparen:    p.parseCall,
		Ternary:   p.parseTernary,
		Eq:        p.parseInfix,
		Ne:        p.parseInfix,
		Lt:        p.parseInfix,
		Le:        p.parseInfix,
		Gt:        p.parseInfix,
		Ge:        p.parseInfix,
		And:       p.parseInfix,
		Or:        p.parseInfix,
	}
	p.next()
	p.next()
	return p.Parse()
}

func (p *parser) Parse() (Expression, error) {
	var s script
	for !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		s.list = append(s.list, e)
		switch p.curr.Type {
		case EOL:
			p.next()
		case EOF:
		default:
			return nil, fmt.Errorf("syntax error! missing eol")
		}
	}
	var e Expression
	switch len(s.list) {
	case 0:
		return nil, fmt.Errorf("empty script given")
	case 1:
		e = s.list[0]
	default:
		e = s
	}
	return e, nil
}

func (p *parser) parse(pow int) (Expression, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("prefix: %s can not be parsed", p.curr)
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for (p.curr.Type != EOL || p.curr.Type != EOF) && pow < powers.Get(p.curr.Type) {
		fn, ok := p.infix[p.curr.Type]
		if !ok {
			return nil, fmt.Errorf("infix: %s can not be parsed", p.curr)
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *parser) parseKeyword() (Expression, error) {
	switch p.curr.Literal {
	case kwIf:
		return p.parseIf()
	case kwWhile:
		return p.parseWhile()
	case kwBreak:
		return p.parseBreak()
	case kwContinue:
		return p.parseContinue()
	case kwReturn:
		return p.parseReturn()
	default:
		return nil, fmt.Errorf("%s: keyword not implemented", p.curr.Literal)
	}
}

func (p *parser) parseBlock() (Expression, error) {
	if p.curr.Type != Lcurly {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	var list []Expression
	for p.curr.Type != Rcurly && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
		if p.curr.Type != EOL {
			return nil, fmt.Errorf("syntax error! missing eol")
		}
		p.next()
	}
	if p.curr.Type != Rcurly {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	switch len(list) {
	case 1:
		return list[0], nil
	default:
		return script{list: list}, nil
	}
}

func (p *parser) parseIf() (Expression, error) {
	p.next()
	if p.curr.Type != Lparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()

	var (
		expr test
		err  error
	)
	expr.cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	expr.csq, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.curr.Type == Keyword && p.curr.Literal == kwElse {
		p.next()
		switch p.curr.Type {
		case Lcurly:
			expr.alt, err = p.parseBlock()
		case Keyword:
			expr.alt, err = p.parseKeyword()
		default:
		}
	}
	if p.curr.Type != EOL && p.curr.Type != EOF {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	return expr, nil
}

func (p *parser) parseWhile() (Expression, error) {
	p.next()
	if p.curr.Type != Lparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()

	var (
		expr while
		err  error
	)
	expr.cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	expr.body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.curr.Type != EOL && p.curr.Type != EOF {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	return expr, nil
}

func (p *parser) parseReturn() (Expression, error) {
	p.next()
	return nil, nil
}

func (p *parser) parseBreak() (Expression, error) {
	p.next()
	return nil, nil
}

func (p *parser) parseContinue() (Expression, error) {
	p.next()
	return nil, nil
}

func (p *parser) parseTernary(left Expression) (Expression, error) {
	var err error
	expr := test{
		cdt: left,
	}
	p.next()
	if expr.csq, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	if p.curr.Type != Alt {
		return nil, fmt.Errorf("syntax error!")
	}
	p.next()

	if expr.alt, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *parser) parseAssign(left Expression) (Expression, error) {
	v, ok := left.(variable)
	if !ok {
		return nil, fmt.Errorf("syntax error!")
	}
	op := p.curr.Type
	p.next()
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr := assign{
		ident: v.ident,
		right: right,
	}
	if op != Assign {
		switch op {
		case AddAssign:
			op = Add
		case SubAssign:
			op = Sub
		case MulAssign:
			op = Mul
		case DivAssign:
			op = Div
		case ModAssign:
			op = Mod
		default:
			return nil, fmt.Errorf("invalid compound assignment operator")
		}
		expr.right = binary{
			op:    op,
			left:  left,
			right: right,
		}
	}
	return expr, nil
}

func (p *parser) parseInfix(left Expression) (Expression, error) {
	expr := binary{
		op:   p.curr.Type,
		left: left,
	}
	pow := powers.Get(p.curr.Type)
	p.next()
	right, err := p.parse(pow)
	if err != nil {
		return nil, err
	}
	expr.right = right
	return expr, nil
}

func (p *parser) parsePrefix() (Expression, error) {
	var expr Expression
	switch p.curr.Type {
	case Sub, Not:
		op := p.curr.Type
		p.next()

		right, err := p.parse(powPrefix)
		if err != nil {
			return nil, err
		}
		expr = unary{
			op:    op,
			right: right,
		}
	case Literal:
		expr = createLiteral(p.curr.Literal)
		p.next()
	case Number:
		n, err := strconv.ParseFloat(p.curr.Literal, 64)
		if err != nil {
			return nil, err
		}
		expr = createNumber(n)
		p.next()
	case Ident:
		expr = createVariable(p.curr.Literal)
		p.next()
	case Boolean:
		b, err := strconv.ParseBool(p.curr.Literal)
		if err != nil {
			return nil, err
		}
		expr = createBoolean(b)
		p.next()
	default:
		return nil, fmt.Errorf("unuspported token: %s", p.curr)
	}
	return expr, nil
}

func (p *parser) parseCall(left Expression) (Expression, error) {
	v, ok := left.(variable)
	if !ok {
		return nil, fmt.Errorf("syntax error! try to call non function")
	}
	p.next()
	expr := call{
		ident: v.ident,
	}
	for p.curr.Type != Rparen && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		expr.args = append(expr.args, e)
		switch p.curr.Type {
		case Comma:
			p.next()
		case Rparen:
		default:
			return nil, fmt.Errorf("syntax error! missing comma")
		}
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("syntax error! missing closing )")
	}
	p.next()
	return expr, nil
}

func (p *parser) parseGroup() (Expression, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("syntax error: missing closing )")
	}
	p.next()
	return expr, nil
}

func (p *parser) done() bool {
	return p.curr.Type == EOF
}

func (p *parser) next() {
	p.curr = p.peek
	p.peek = p.lex.Lex()
}
