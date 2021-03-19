package main

import "fmt"

type RuntimeError string

func (e RuntimeError) Error() string {
	return string(e)
}

func runtimeErr(t *tokenObj, msg string) error {
	panic(RuntimeError(
		fmt.Sprintf("[line %v] runtime error: %v", t.line, msg)))
}

// ------------------------------------------
// env

// env contains bindings for variables.
type env struct {
	values    map[string]value
	enclosing *env
}

func NewEnv(enclosing *env) *env {
	return &env{make(map[string]value), enclosing}
}

func (e *env) define(name string, v value) {
	e.values[name] = v
}

func (e *env) get(name *tokenObj) value {
	if v, ok := e.values[name.lexeme]; ok {
		return v
	}
	if e.enclosing != nil {
		return e.enclosing.get(name)
	}

	runtimeErr(name, "undefined variable '"+name.lexeme+"'")
	return nil
}

func (e *env) assign(name *tokenObj, v value) {
	if _, ok := e.values[name.lexeme]; ok {
		e.values[name.lexeme] = v
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

// This var being a global var is a dirty hack.
// TODO: fix this hack into some local var.
var curenv *env

func interpret(stmt []Stmt) (err error) {
	curenv = NewEnv(nil) // root env has no enclosure

	defer func() {
		if e := recover(); e != nil {
			err = e.(RuntimeError)
		}
	}()
	for _, s := range stmt {
		s.execute()
	}
	return nil
}

func (e *BinaryExpr) eval() value {
	switch e.operator.tok {
	case Plus:
		x := e.left.eval()
		xval, xok := x.(float64)
		if xok {
			y := e.right.eval()
			yval, yok := y.(float64)
			if yok {
				return xval + yval
			}
			runtimeErr(e.operator, "expected number as right operand")
		}
		if xval, xok := x.(string); xok {
			if yval, yok := e.right.eval().(string); yok {
				return xval + yval
			}
			runtimeErr(e.operator, "expected string as right operand")
		}
		runtimeErr(e.operator, "operands must be two numbers or two strings")
	case Minus:
		xval, yval := e.evalFloats()
		return xval - yval
	case Slash:
		xval, yval := e.evalFloats()
		if yval == 0 {
			runtimeErr(e.operator, "division by zero")
		}
		return xval / yval
	case Star:
		xval, yval := e.evalFloats()
		return xval * yval
	case Greater:
		xval, yval := e.evalFloats()
		return xval > yval
	case GreaterEqual:
		xval, yval := e.evalFloats()
		return xval >= yval
	case Less:
		xval, yval := e.evalFloats()
		return xval < yval
	case LessEqual:
		xval, yval := e.evalFloats()
		return xval <= yval
	case EqualEqual:
		return e.equal()
	case BangEqual:
		return !e.equal()
	}
	return nil // Unreachable?
}

func (e *BinaryExpr) evalFloats() (float64, float64) {
	x, ok := e.left.eval().(float64)
	if !ok {
		runtimeErr(e.operator, "left operand must be a number")
	}
	y, ok := e.right.eval().(float64)
	if !ok {
		runtimeErr(e.operator, "left operand must be a number")
	}
	return x, y
}

func (e *BinaryExpr) equal() bool {
	x := e.left.eval()
	y := e.right.eval()
	if x == nil && y == nil {
		return true
	}
	if x == nil {
		return false
	}
	return x == y
}

func (e *GroupingExpr) eval() value {
	return e.e.eval()
}

func (e *LiteralExpr) eval() value {
	return e.value
}

func (e *UnaryExpr) eval() value {
	val := e.right.eval()
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

func (e *VarExpr) eval() value {
	return curenv.get(e.name)
}

func (e *AssignExpr) eval() value {
	v := e.value.eval()
	curenv.assign(e.name, v)
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

func (s *ExprStmt) execute() {
	s.expression.eval()
}

func (s *PrintStmt) execute() {
	v := s.expression.eval()
	fmt.Println(v)
}

func (s *VarStmt) execute() {
	var v value
	if s.init != nil {
		v = s.init.eval()
	}
	curenv.define(s.name.lexeme, v)
}

func (s *BlockStmt) execute() {
	execBlock(s.list, NewEnv(curenv))
}

func execBlock(list []Stmt, e *env) {
	saved := curenv
	curenv = e
	defer func() { curenv = saved }()
	for _, s := range list {
		s.execute()
	}
}
