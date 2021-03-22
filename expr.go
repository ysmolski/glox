package main

type (
	value interface{}

	Expr interface {
		aExpr()
		eval(*Env) value
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

	LogicalExpr struct {
		operator    *tokenObj
		left, right Expr
		expr
	}

	UnaryExpr struct {
		operator *tokenObj
		right    Expr
		expr
	}

	VarExpr struct {
		name *tokenObj
		expr
	}

	AssignExpr struct {
		name  *tokenObj
		value Expr
		expr
	}
)

func (*expr) aExpr()          {}
func (*expr) eval(*Env) value { return nil }

type (
	Stmt interface {
		aStmt()
		execute(*Env)
	}

	stmt struct{}

	ExprStmt struct {
		expression Expr
		stmt
	}

	PrintStmt struct {
		expression Expr
		stmt
	}

	VarStmt struct {
		name *tokenObj
		init Expr
		stmt
	}

	BlockStmt struct {
		list []Stmt
		stmt
	}

	IfStmt struct {
		condition Expr
		a, b      Stmt
		stmt
	}

	WhileStmt struct {
		condition Expr
		body      Stmt
		stmt
	}
)

func (*stmt) aStmt()       {}
func (*stmt) execute(*Env) {}

// func printAST(e Expr) string {
// 	switch o := e.(type) {
// 	case *BinaryExpr:
// 		return fmt.Sprintf("(%v %v %v)",
// 			o.operator.tok, printAST(o.left), printAST(o.right))
// 	case *TernaryExpr:
// 		return fmt.Sprintf("(%v %v %v %v)",
// 			o.operator.tok, printAST(o.op1), printAST(o.op2), printAST(o.op3))
// 	case *UnaryExpr:
// 		return fmt.Sprintf("(%v %v)",
// 			o.operator.tok, printAST(o.right))
// 	case *GroupingExpr:
// 		return fmt.Sprintf("(group %v)", printAST(o.e))
// 	case *LiteralExpr:
// 		return fmt.Sprintf("%v", o.value)
// 	default:
// 		panic("unexpected type of expr")
// 	}
// }
