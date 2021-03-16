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
		os.Exit(65)
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
	for _, t := range tokens {
		fmt.Println("token ", t)
	}
}

func report(line int, msg string) {
	reportDet(line, "", msg)
}

func reportDet(line int, where, msg string) {
	fmt.Printf("[line %v] Error%v: %v", line, where, msg)
	hadError = true
}
