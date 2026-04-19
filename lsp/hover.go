package main

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/sachinbhankhar/noname/evaluator"
	"github.com/sachinbhankhar/noname/lexer"
	"github.com/sachinbhankhar/noname/object"
	"github.com/sachinbhankhar/noname/parser"
)

func textDocumentHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := string(params.TextDocument.URI)
	content, ok := documents[uri]
	if !ok {
		return nil, nil
	}

	// Step 1: find the word the user is hovering over.
	word := wordAt(content, int(params.Position.Line), int(params.Position.Character))
	if word == "" {
		return nil, nil
	}

	// Step 2: check if it's a builtin first.
	if desc, ok := builtinDescriptions[word]; ok {
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: fmt.Sprintf("**builtin** `%s`\n\n%s", word, desc),
			},
		}, nil
	}

	// Step 3: evaluate the whole file and look up the identifier in the env.
	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, nil
	}

	env := object.NewEnvironment()
	evaluator.Eval(program, env)

	obj, ok := env.Get(word)
	if !ok {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: fmt.Sprintf("**%s** `%s`", strings.ToLower(string(obj.Type())), obj.Inspect()),
		},
	}, nil
}

// wordAt extracts the identifier sitting at line/col in content.
// Lines and cols are 0-based (LSP convention).
func wordAt(content string, line, col int) string {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return ""
	}
	src := lines[line]
	if col >= len(src) {
		return ""
	}

	// Walk backward to find start of word.
	start := col
	for start > 0 && isIdentChar(src[start-1]) {
		start--
	}

	// Walk forward to find end of word.
	end := col
	for end < len(src) && isIdentChar(src[end]) {
		end++
	}

	return src[start:end]
}

func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || (ch >= '0' && ch <= '9')
}

// builtinDescriptions are shown when hovering over builtin function names.
var builtinDescriptions = map[string]string{
	"print": "Print one or more values to stdout, separated by newlines.\n\n`print(value, ...)`",
	"puts":  "Print one or more values to stdout, separated by newlines.\n\n`puts(value, ...)`",
	"len":  "Return the length of a string.\n\n`len(string) -> integer`",
	"str":  "Convert any value to its string representation.\n\n`str(value) -> string`",
	"type": "Return the type name of a value as a string.\n\n`type(value) -> string`",
}
