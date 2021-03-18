package main

import "fmt"

type (
	value interface{}

	Expr interface {
		aExpr()
		eval() value
	}

	expr struct{}

	BinaryExpr struct {
		operator    *tokenObj
		left, right Expr
		expr
	}

	GroupingExpr struct {
		e Expr
		expr
	}

	LiteralExpr struct {
		value interface{}
		expr
	}

	TernaryExpr struct {
		operator      *tokenObj
		op1, op2, op3 Expr
		expr
	}

	UnaryExpr struct {
		operator *tokenObj
		right    Expr
		expr
	}
)

func (*expr) aExpr()      {}
func (*expr) eval() value { return nil }

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
			panic(runtimeErr(e.operator, "expected number as right operand"))
		}
		if xval, xok := x.(string); xok {
			if yval, yok := e.right.eval().(string); yok {
				return xval + yval
			}
			panic(runtimeErr(e.operator, "expected string as right operand"))
		}
		panic(runtimeErr(e.operator, "operands must be two numbers or two strings"))
	case Minus:
		xval, yval := e.evalFloats()
		return xval - yval
	case Slash:
		xval, yval := e.evalFloats()
		if yval == 0 {
			panic(runtimeErr(e.operator, "division by zero"))
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
		panic(runtimeErr(e.operator, "left operand must be a number"))
	}
	y, ok := e.right.eval().(float64)
	if !ok {
		panic(runtimeErr(e.operator, "left operand must be a number"))
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

func printAST(e Expr) string {
	switch o := e.(type) {
	case *BinaryExpr:
		return fmt.Sprintf("(%v %v %v)",
			o.operator.tok, printAST(o.left), printAST(o.right))
	case *TernaryExpr:
		return fmt.Sprintf("(%v %v %v %v)",
			o.operator.tok, printAST(o.op1), printAST(o.op2), printAST(o.op3))
	case *UnaryExpr:
		return fmt.Sprintf("(%v %v)",
			o.operator.tok, printAST(o.right))
	case *GroupingExpr:
		return fmt.Sprintf("(group %v)", printAST(o.e))
	case *LiteralExpr:
		return fmt.Sprintf("%v", o.value)
	default:
		panic("unexpected type of expr")
	}
}
