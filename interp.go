package main

import (
	"errors"
	"fmt"
	"time"
)

type RuntimeError string

func (e RuntimeError) Error() string {
	return string(e)
}

func runtimeErr(t *tokenObj, msg string) error {
	panic(RuntimeError(
		fmt.Sprintf("[line %v] runtime error: %v", t.line, msg)))
}

type ReturnHack value
type BreakErr struct{ t *tokenObj }
type ContinueErr struct{ t *tokenObj }

type Callable interface {
	arity() int
	call(*Env, []value) value
}

// ------------------------------------------
// env

// env contains bindings for variables.
type Env struct {
	values map[string]value

	// init means that variable was properly initialized
	init map[string]bool

	enclosing *Env
	globals   *Env // always points to the root of enclosures
}

func NewEnv(enclosing *Env) *Env {
	e := &Env{make(map[string]value), make(map[string]bool), enclosing, nil}
	if enclosing == nil {
		// means that this created env is the root, that is global env
		e.globals = e
	} else {
		e.globals = enclosing.globals
	}
	return e
}

func (e *Env) defineInit(name string, v value) {
	e.values[name] = v
	e.init[name] = true
}

func (e *Env) define(name string) {
	e.values[name] = nil
}

func (e *Env) get(name *tokenObj) value {
	if v, ok := e.values[name.lexeme]; ok {
		if _, ok := e.init[name.lexeme]; !ok {
			runtimeErr(name, "variable '"+name.lexeme+"' should be initialized first")
		}
		return v
	}
	if e.enclosing != nil {
		return e.enclosing.get(name)
	}

	runtimeErr(name, "undefined variable '"+name.lexeme+"'")
	return nil
}

func (e *Env) assign(name *tokenObj, v value) {
	if _, ok := e.values[name.lexeme]; ok {
		e.values[name.lexeme] = v
		e.init[name.lexeme] = true
		return
	}
	if e.enclosing != nil {
		e.enclosing.assign(name, v)
		return
	}

	runtimeErr(name, "undefined variable '"+name.lexeme+"'")
	return
}

// ------------------------------------------
// interpret

func interpret(stmt []Stmt, env *Env) (err error) {
	env.defineInit("clock", clockFn{})
	defer func() {
		if e := recover(); e != nil {
			if b, ok := e.(BreakErr); ok {
				fmt.Printf("b.t = %+v\n", b.t.line)
				s := fmt.Sprintf("expected a while loop to break from at line %v ", b.t.line)
				err = errors.New(s)
				return
			}
			err = e.(RuntimeError)
		}
	}()
	for _, s := range stmt {
		s.execute(env)
	}
	return nil
}

// ------------------------------------------
// clockFn

type clockFn struct{}

func (c clockFn) arity() int {
	return 0
}

func (c clockFn) call(_ *Env, _ []value) value {
	return float64(time.Now().UnixNano())
}

// ------------------------------------------
// Function

type FunObj struct {
	decl    *FunStmt
	closure *Env
}

func (f *FunObj) arity() int {
	return len(f.decl.params)
}

func (f *FunObj) call(e *Env, args []value) (v value) {
	env := NewEnv(f.closure)
	for i, p := range f.decl.params {
		env.defineInit(p.lexeme, args[i])
	}

	defer func() {
		if e := recover(); e != nil {
			// return whatever value is being panicked at us from return stmt
			v = e.(ReturnHack)
		}
	}()
	execBlock(f.decl.body, env)
	return nil
}

func (f *FunObj) String() string {
	return fmt.Sprintf("<fn %v>", f.decl.name.lexeme)
}

// ------------------------------------------
// Expression Eval

func (e *BinaryExpr) eval(env *Env) value {
	switch e.operator.tok {
	case Plus:
		x := e.left.eval(env)
		xval, xok := x.(float64)
		if xok {
			y := e.right.eval(env)
			yval, yok := y.(float64)
			if yok {
				return xval + yval
			}
			runtimeErr(e.operator, "expected number as right operand")
		}
		if xval, xok := x.(string); xok {
			if yval, yok := e.right.eval(env).(string); yok {
				return xval + yval
			}
			runtimeErr(e.operator, "expected string as right operand")
		}
		runtimeErr(e.operator, "operands must be two numbers or two strings")
	case Minus:
		xval, yval := e.evalFloats(env)
		return xval - yval
	case Slash:
		xval, yval := e.evalFloats(env)
		if yval == 0 {
			runtimeErr(e.operator, "division by zero")
		}
		return xval / yval
	case Star:
		xval, yval := e.evalFloats(env)
		return xval * yval
	case Greater:
		xval, yval := e.evalFloats(env)
		return xval > yval
	case GreaterEqual:
		xval, yval := e.evalFloats(env)
		return xval >= yval
	case Less:
		xval, yval := e.evalFloats(env)
		return xval < yval
	case LessEqual:
		xval, yval := e.evalFloats(env)
		return xval <= yval
	case EqualEqual:
		return e.equal(env)
	case BangEqual:
		return !e.equal(env)
	}
	return nil // Unreachable?
}

func (e *BinaryExpr) evalFloats(env *Env) (float64, float64) {
	x, ok := e.left.eval(env).(float64)
	if !ok {
		runtimeErr(e.operator, "left operand must be a number")
	}
	y, ok := e.right.eval(env).(float64)
	if !ok {
		runtimeErr(e.operator, "left operand must be a number")
	}
	return x, y
}

func (e *BinaryExpr) equal(env *Env) bool {
	x := e.left.eval(env)
	y := e.right.eval(env)
	if x == nil && y == nil {
		return true
	}
	if x == nil {
		return false
	}
	return x == y
}

func (e *CallExpr) eval(env *Env) value {
	callee := e.callee.eval(env)
	args := make([]value, 0)
	for _, a := range e.args {
		args = append(args, a.eval(env))
	}
	if fn, ok := callee.(Callable); ok {
		if len(args) != fn.arity() {
			runtimeErr(e.paren,
				fmt.Sprintf("expected %v arguments but got %v", fn.arity(), len(args)))
		}
		return fn.call(env, args)
	} else {
		err := fmt.Sprintf("'%v' is not a function or class", callee)
		runtimeErr(e.paren, err)
		return nil
	}
}

func (e *GroupingExpr) eval(env *Env) value {
	return e.e.eval(env)
}

func (e *LiteralExpr) eval(env *Env) value {
	return e.value
}

func (e *LogicalExpr) eval(env *Env) value {
	left := e.left.eval(env)
	if e.operator.tok == Or {
		if isTruthy(left) {
			return left
		}
	} else {
		if !isTruthy(left) {
			return left
		}
	}
	return e.right.eval(env)
}

func (e *UnaryExpr) eval(env *Env) value {
	val := e.right.eval(env)
	switch e.operator.tok {
	case Minus:
		f, ok := val.(float64)
		if !ok {
			// TODO: handle this as error
			panic("not a float")
		}
		return -f
	case Bang:
		return !isTruthy(val)
	}
	// unreachable?
	return nil
}

func (e *VarExpr) eval(env *Env) value {
	return env.get(e.name)
}

func (e *AssignExpr) eval(env *Env) value {
	v := e.value.eval(env)
	env.assign(e.name, v)
	return v
}

// false and nil are the only falsey values
func isTruthy(v value) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

// --------------------------------------------------------
// Statements

func (s *ExprStmt) execute(env *Env) {
	s.expression.eval(env)
}

func (s *FunStmt) execute(env *Env) {
	fn := &FunObj{decl: s, closure: NewEnv(env)}
	env.defineInit(s.name.lexeme, fn)
}

func (s *PrintStmt) execute(env *Env) {
	v := s.expression.eval(env)
	fmt.Println(v)
}

func (s *VarStmt) execute(env *Env) {
	// make distinction between uninitialized value and nil-value
	if s.init != nil {
		v := s.init.eval(env)
		env.defineInit(s.name.lexeme, v)
	} else {
		env.define(s.name.lexeme)
	}
}

func (s *BlockStmt) execute(env *Env) {
	execBlock(s.list, NewEnv(env))
}

func execBlock(list []Stmt, env *Env) {
	for _, s := range list {
		s.execute(env)
	}
}

func (s *IfStmt) execute(env *Env) {
	if isTruthy(s.condition.eval(env)) {
		s.block1.execute(env)
	} else if s.block2 != nil {
		s.block2.execute(env)
	}
}

func (s *ReturnStmt) execute(env *Env) {
	var v value
	if s.value != nil {
		v = s.value.eval(env)
	}
	// Ugly hack, panic to unwind the stack back to the call
	panic(ReturnHack(v))
}

func (s *BreakStmt) execute(env *Env) {
	panic(BreakErr{t: s.keyword})
}

func (s *ContinueStmt) execute(env *Env) {
	panic(ContinueErr{t: s.keyword})
}

func (s *WhileStmt) execute(env *Env) {
	for !s.isDone(env) {
	}
}

// isDone returns false when the loop was continued,
// when loop is done returns true
func (s *WhileStmt) isDone(env *Env) (done bool) {
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case ContinueErr:
				done = false
				return
			case BreakErr:
				done = true
				return
			default:
				panic(e)
			}
		}
	}()
	for isTruthy(s.condition.eval(env)) {
		s.body.execute(env)
	}
	return true
}
