package main

import "fmt"

type (
	Expr interface {
		aExpr()
	}

	expr struct{}

	BinaryExpr struct {
		operator    *tokenObj
		left, right Expr
		expr
	}

	TernaryExpr struct {
		operator      *tokenObj
		op1, op2, op3 Expr
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

	UnaryExpr struct {
		operator *tokenObj
		right    Expr
		expr
	}
)

func (*expr) aExpr() {}

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
