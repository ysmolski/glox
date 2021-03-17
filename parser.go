package main

import (
	"errors"
	"fmt"
)

// expression     → equality ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary
//                | primary ;
// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")" ;

type Parser struct {
	tokens  []*tokenObj
	current int
}

func (p *Parser) match(toks ...token) bool {
	for _, t := range toks {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) advance() *tokenObj {
	if !p.atEnd() {
		p.current++
	}
	return p.prev()
}

func (p *Parser) atEnd() bool {
	return p.peek().tok == EOF
}

func (p *Parser) peek() *tokenObj {
	return p.tokens[p.current]
}

func (p *Parser) prev() *tokenObj {
	return p.tokens[p.current-1]
}

func (p *Parser) check(tok token) bool {
	if p.atEnd() {
		return false
	}
	fmt.Printf("p.peek() = %+v\n", p.peek())
	return p.peek().tok == tok
}

// ---------------------------------------------------------
//

func (p *Parser) parse() Expr {
	defer func() {
		e := recover()
		if e != nil {
			fmt.Println("Recovered in parse()", e)
		}
	}()
	return p.expression()
}

func (p *Parser) expression() Expr {
	return p.equality()
}

func (p *Parser) equality() Expr {
	expr := p.comparison()
	for p.match(Bang, BangEqual) {
		op := p.prev()
		right := p.comparison()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()
	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		op := p.prev()
		right := p.term()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// term →  factor ( ( "-" | "+" ) factor )* ;
func (p *Parser) term() Expr {
	expr := p.factor()
	for p.match(Plus, Minus) {
		op := p.prev()
		right := p.factor()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// factor         → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser) factor() Expr {
	expr := p.unary()
	for p.match(Slash, Star) {
		op := p.prev()
		right := p.unary()
		expr = &BinaryExpr{operator: op, left: expr, right: right}
	}
	return expr
}

// unary          → ( "!" | "-" ) unary
//                | primary ;
func (p *Parser) unary() Expr {
	if p.match(Bang, Minus) {
		op := p.prev()
		right := p.unary()
		return &UnaryExpr{operator: op, right: right}
	}
	return p.primary()
}

// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")" ;
func (p *Parser) primary() Expr {
	switch {
	case p.match(False):
		return &LiteralExpr{value: false}
	case p.match(True):
		return &LiteralExpr{value: true}
	case p.match(Nil):
		return &LiteralExpr{value: nil}
	case p.match(Number, String):
		return &LiteralExpr{value: p.prev().literal}
	case p.match(LeftParen):
		expr := p.expression()
		p.consume(RightParen, "Expect ')' after expression")
		return &GroupingExpr{e: expr}
	}
	panic(p.err(p.peek(), "expect expression"))
}

func (p *Parser) consume(expected token, msg string) *tokenObj {
	if p.check(expected) {
		return p.advance()
	}
	panic(p.err(p.peek(), msg))
}

func (p *Parser) err(t *tokenObj, msg string) error {
	reportToken(t, msg)
	return errors.New("parsing error")
}

func (p *Parser) sync() {
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
