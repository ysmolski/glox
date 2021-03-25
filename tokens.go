package main

import "fmt"

type token uint

//go:generate stringer -type token -linecomment tokens.go

const (
	// single character tokens
	_          token = iota
	LeftParen        // (
	RightParen       // )
	LeftBrace        // {
	RightBrace       // }
	Comma            // ,
	Dot              // .
	Minus            // -
	Plus             // +
	Semicolon        // ;
	Colon            // :
	Question         // ?
	Slash            // /
	Star             // *

	Bang         // !
	BangEqual    // !=
	Equal        // =
	EqualEqual   // ==
	Greater      // >
	GreaterEqual // >=
	Less         // <
	LessEqual    // <=

	Identifier // ident
	String     // string
	Number     // number

	And      // and
	Break    // break
	Class    // class
	Continue // continue
	Else     // else
	False    // false
	Fun      // fun
	For      // for
	If       // if
	Nil      // nil
	Or       // or
	Print    // print
	Return   // return
	Super    // super
	This     // this
	True     // true
	Var      // var
	While    // while

	EOF // eof
)

type tokenObj struct {
	tok     token
	lexeme  string
	line    int
	literal interface{}
}

func (t *tokenObj) String() string {
	return fmt.Sprintf("token: %v lex: %v lit: %v", t.tok, t.lexeme, t.literal)
}
