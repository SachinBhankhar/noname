package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/sachinbhankhar/noname/evaluator"
	"github.com/sachinbhankhar/noname/lexer"
	"github.com/sachinbhankhar/noname/object"
	"github.com/sachinbhankhar/noname/parser"
	"github.com/sachinbhankhar/noname/repl"
)

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s\n", user.Username)
	repl.Start(os.Stdin, os.Stdout)
}

func runFile(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil && result.Type() == object.ERROR_OBJ {
		fmt.Fprintln(os.Stderr, result.Inspect())
		os.Exit(1)
	}
}
