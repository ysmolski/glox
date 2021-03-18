package main

import (
	"bufio"
	"errors"
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
	tokens := scanner.scan()
	// for _, t := range tokens {
	// 	fmt.Println("token ", t)
	// }
	p := &Parser{tokens, 0}
	expr := p.parse()

	if hadError {
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

func report(line int, msg string) {
	reportLoc(line, "", msg)
}

func reportToken(t *tokenObj, msg string) {
	if t.tok == EOF {
		reportLoc(t.line, " at end", msg)
	} else {
		reportLoc(t.line, " at '"+t.lexeme+"'", msg)
	}
}
func reportLoc(line int, where, msg string) {
	fmt.Printf("[line %v] error%v: %v\n", line, where, msg)
	hadError = true
}

func runtimeErr(t *tokenObj, msg string) error {
	return errors.New(fmt.Sprintf("[line %v] runtime error: %v",
		t.line, msg))
}
