package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/sachinbhankhar/noname/ast"
	"github.com/sachinbhankhar/noname/lexer"
	"github.com/sachinbhankhar/noname/parser"
)

func textDocumentCompletion(ctx *glsp.Context, params *protocol.CompletionParams) (any, error) {
	uri := string(params.TextDocument.URI)
	content, ok := documents[uri]
	if !ok {
		return nil, nil
	}

	items := []protocol.CompletionItem{}

	// 1. Keywords
	items = append(items, keywordCompletions()...)

	// 2. Builtins
	items = append(items, builtinCompletions()...)

	// 3. User-defined variables collected from the AST
	items = append(items, variableCompletions(content)...)

	return items, nil
}

// keywordCompletions returns completion items for all language keywords.
func keywordCompletions() []protocol.CompletionItem {
	kind := protocol.CompletionItemKindKeyword
	keywords := []string{"let", "fn", "if", "else", "return", "true", "false"}

	items := []protocol.CompletionItem{}
	for _, kw := range keywords {
		word := kw // capture loop var
		items = append(items, protocol.CompletionItem{
			Label: word,
			Kind:  &kind,
		})
	}
	return items
}

// builtinCompletions returns completion items with documentation for each builtin.
func builtinCompletions() []protocol.CompletionItem {
	kind := protocol.CompletionItemKindFunction
	items := []protocol.CompletionItem{}

	for name, desc := range builtinDescriptions {
		n, d := name, desc // capture loop vars
		items = append(items, protocol.CompletionItem{
			Label:         n,
			Kind:          &kind,
			Documentation: d,
		})
	}
	return items
}

// variableCompletions parses the file and returns every let-bound name.
func variableCompletions(content string) []protocol.CompletionItem {
	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()

	kind := protocol.CompletionItemKindVariable
	items := []protocol.CompletionItem{}
	seen := map[string]bool{}

	for _, stmt := range program.Statements {
		letStmt, ok := stmt.(*ast.LetStatement)
		if !ok {
			continue
		}
		name := letStmt.Name.Value
		if seen[name] {
			continue
		}
		seen[name] = true
		items = append(items, protocol.CompletionItem{
			Label: name,
			Kind:  &kind,
		})
	}
	return items
}
