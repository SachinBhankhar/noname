package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/sachinbhankhar/noname/evaluator"
	"github.com/sachinbhankhar/noname/lexer"
	"github.com/sachinbhankhar/noname/object"
	"github.com/sachinbhankhar/noname/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				fmt.Fprintf(out, "\t%s\n", msg)
			}
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			fmt.Fprintln(out, evaluated.Inspect())
		}
	}
}
