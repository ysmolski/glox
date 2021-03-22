package main

import "fmt"

// Recursive-descent parser
//
// program        -> declaration* EOF ;
//
// declaration    -> varDecl
//                 | statement ;
//
// varDecl        -> "var" IDENTIFIER ( "=" expression )? ";" ;
//
// statement      -> exprStmt
//                 | printStmt
//				   | block ;
//
// exprStmt       -> expression ";" ;
// printStmt      -> "print" expression ";" ;
// block		  -> "{" declaration* "}" ;
//
// expression     -> assignment ;
// assignment     -> IDENTIFIER "=" assignment
//				   | equality ;
// equality       -> comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     -> term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           -> factor ( ( "-" | "+" ) factor )* ;
// factor         -> unary ( ( "/" | "*" ) unary )* ;
// unary          -> ( "!" | "-" ) unary
//                 | primary ;
// primary        -> NUMBER | STRING | "true" | "false" | "nil"
//                 | "(" expression ")"
//                 | IDENTIFIER ;
//

type parser struct {
	tokens  []*tokenObj
	current int
	errs    []error
}

func NewParser(tokens []*tokenObj) *parser {
	p := &parser{tokens, 0, make([]error, 0)}
	return p
}

// match advances pointer to the next token if current token matches
// any of toks and returns true
func (p *parser) match(toks ...token) bool {
	for _, t := range toks {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) advance() *tokenObj {
	if !p.atEnd() {
		p.current++
	}
	return p.prev()
}

func (p *parser) atEnd() bool {
	return p.peek().tok == EOF
}

func (p *parser) peek() *tokenObj {
	return p.tokens[p.current]
}

func (p *parser) prev() *tokenObj {
	return p.tokens[p.current-1]
}

func (p *parser) check(tok token) bool {
	if p.atEnd() {
		return false
	}
	// fmt.Printf("p.peek() = %+v\n", p.peek())
	return p.peek().tok == tok
}

func (p *parser) consume(expected token, msg string) *tokenObj {
	if p.check(expected) {
		return p.advance()
	}
	p.perror(p.peek(), msg)
	return nil
}

type ParsingError string

func (e ParsingError) Error() string {
	return string(e)
}

func (p *parser) perror(t *tokenObj, msg string) {
	e := ParsingError(errorAtToken(t, msg))
	p.errs = append(p.errs, e)
	panic(e)
}

func (p *parser) yerror(t *tokenObj, msg string) {
	e := ParsingError(errorAtToken(t, msg))
	p.errs = append(p.errs, e)
}

func (p *parser) sync() {
	fmt.Println("sync")
	p.advance()
	for !p.atEnd() {
		if p.prev().tok == Semicolon {
			return
		}
		switch p.peek().tok {
		case Class, Fun, Var, For, If, While, Print, Return:
			return
		}
		p.advance()
	}
}

// ---------------------------------------------------------
//

// parse returns an AST of parsed tokens, if it cannot parse then it returns
// the error.
func (p *parser) parse() (s []Stmt, errs []error) {
	s = make([]Stmt, 0)
	for !p.atEnd() {
		s = append(s, p.declaration())
	}

	return s, p.errs
}

func (p *parser) declaration() (s Stmt) {
	defer func() {
		if e := recover(); e != nil {
			_ = e.(ParsingError) // Panic for other errors
			p.sync()
			s = nil
		}
	}()
	if p.match(Var) {
		return p.varDecl()
	}
	return p.statement()
}

func (p *parser) varDecl() Stmt {
	name := p.consume(Identifier, "expected variable name")
	var init Expr

	if p.match(Equal) {
		init = p.expression()
	}
	p.consume(Semicolon, "expected ';' after variable declaration")
	return &VarStmt{name: name, init: init}
}

func (p *parser) statement() Stmt {
	if p.match(Print) {
		return p.printStatement()
	}
	if p.match(LeftBrace) {
		return &BlockStmt{list: p.block()}
	}
	return p.exprStatement()
}

func (p *parser) printStatement() Stmt {
	e := p.expression()
	p.consume(Semicolon, "expected ';' after expression")
	return &PrintStmt{expression: e}
}

func (p *parser) block() []Stmt {
	list := make([]Stmt, 0)
	for !p.check(RightBrace) && !p.atEnd() {
		list = append(list, p.declaration())
	}
	p.consume(RightBrace, "expected '}' after block")
	return list
}

func (p *parser) exprStatement() Stmt {
	e := p.expression()
	p.consume(Semicolon, "expected ';' after expression")
	return &ExprStmt{expression: e}
}

// expression -> sequential ;
func (p *parser) expression() Expr {
	return p.assignment()
}

func (p *parser) assignment() Expr {
	expr := p.equality()
	if p.match(Equal) {
		equals := p.prev()
		value := p.assignment()
		if ev, ok := expr.(*VarExpr); ok {
			name := ev.name
			return &AssignExpr{name: name, value: value}
		}
		p.yerror(equals, "invalid assignment target")
	}
	return expr
}

// equality -> comparison ( ( "!=" | "==" ) comparison )* ;
func (p *parser) equality() Expr {
	expr := p.comparison()
	for p.match(BangEqual, EqualEqual) {
		op := p.prev()
		right := p.comparison()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// comparison -> term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *parser) comparison() Expr {
	expr := p.term()
	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		op := p.prev()
		right := p.term()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// term ->  factor ( ( "-" | "+" ) factor )* ;
func (p *parser) term() Expr {
	expr := p.factor()
	for p.match(Plus, Minus) {
		op := p.prev()
		right := p.factor()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// factor -> unary ( ( "/" | "*" ) unary )* ;
func (p *parser) factor() Expr {
	expr := p.unary()
	for p.match(Slash, Star) {
		op := p.prev()
		right := p.unary()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// unary -> ( "!" | "-" ) unary
//        | primary ;
func (p *parser) unary() Expr {
	if p.match(Bang, Minus) {
		op := p.prev()
		right := p.unary()
		return &UnaryExpr{operator: op, right: right}
	}
	return p.primary()
}

// primary -> NUMBER | STRING | "true" | "false" | "nil"
//          | "(" expression ")" ;
func (p *parser) primary() Expr {
	switch {
	case p.match(False):
		return &LiteralExpr{value: false}
	case p.match(True):
		return &LiteralExpr{value: true}
	case p.match(Nil):
		return &LiteralExpr{value: nil}
	case p.match(Number, String):
		return &LiteralExpr{value: p.prev().literal}
	case p.match(Identifier):
		return &VarExpr{name: p.prev()}
	case p.match(LeftParen):
		expr := p.expression()
		p.consume(RightParen, "expected enclosing ')' after expression")
		return &GroupingExpr{e: expr}
	}
	p.perror(p.peek(), "expected expression")
	return nil
}
