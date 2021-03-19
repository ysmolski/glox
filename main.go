package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

var hadError = false

func main() {
	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprint(os.Stderr, "usage: glox [script]\n")
		os.Exit(63)
	} else if len(args) == 1 {
		runFile(args[0])
	} else {
		runPrompt()
	}
}

func runFile(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	run(string(data))
	if hadError {
		os.Exit(1)
	}
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		run(line)
		hadError = false
	}
}

func run(source string) {
	scanner := NewScanner(source)
	tokens, err := scanner.scan()
	if err != nil {
		fmt.Println(err)
		hadError = true
		return
	}
	// for _, t := range tokens {
	// 	fmt.Println("token ", t)
	// }

	p := &Parser{tokens, 0}
	expr, err := p.parse()
	if err != nil {
		fmt.Println(err)
		hadError = true
		return
	}
	// fmt.Printf("ast = %+v\n", printAST(expr))

	// catch runtime error, print and bail
	defer func() {
		e := recover()
		if e != nil {
			fmt.Println(e)
			hadError = true
		}
	}()
	val := expr.eval()
	fmt.Println(val)
}

func errorAtToken(t *tokenObj, msg string) string {
	var e string
	if t.tok == EOF {
		e = errorAt(t.line, " at end", msg)
	} else {
		e = errorAt(t.line, " at '"+t.lexeme+"'", msg)
	}
	return e
}

func errorAt(line int, where, msg string) string {
	return fmt.Sprintf("[line %v] error%v: %v", line, where, msg)
}

type RuntimeError string

func (e RuntimeError) Error() string {
	return string(e)
}

func runtimeErr(t *tokenObj, msg string) error {
	return RuntimeError(fmt.Sprintf("[line %v] runtime error: %v",
		t.line, msg))
}
