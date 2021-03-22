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
type Env struct {
	values    map[string]value
	enclosing *Env
}

func NewEnv(enclosing *Env) *Env {
	return &Env{make(map[string]value), enclosing}
}

func (e *Env) define(name string, v value) {
	e.values[name] = v
}

func (e *Env) get(name *tokenObj) value {
	if v, ok := e.values[name.lexeme]; ok {
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
	defer func() {
		if e := recover(); e != nil {
			err = e.(RuntimeError)
		}
	}()
	for _, s := range stmt {
		s.execute(env)
	}
	return nil
}

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

func (e *GroupingExpr) eval(env *Env) value {
	return e.e.eval(env)
}

func (e *LiteralExpr) eval(env *Env) value {
	return e.value
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

func (s *PrintStmt) execute(env *Env) {
	v := s.expression.eval(env)
	fmt.Println(v)
}

func (s *VarStmt) execute(env *Env) {
	var v value
	if s.init != nil {
		v = s.init.eval(env)
	}
	env.define(s.name.lexeme, v)
}

func (s *BlockStmt) execute(env *Env) {
	execBlock(s.list, NewEnv(env))
}

func execBlock(list []Stmt, env *Env) {
	for _, s := range list {
		s.execute(env)
	}
}
