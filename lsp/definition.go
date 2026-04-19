package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/sachinbhankhar/noname/ast"
	"github.com/sachinbhankhar/noname/lexer"
	"github.com/sachinbhankhar/noname/parser"
)

func textDocumentDefinition(ctx *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	uri := string(params.TextDocument.URI)
	content, ok := documents[uri]
	if !ok {
		return nil, nil
	}

	// Find the word under the cursor.
	word := wordAt(content, int(params.Position.Line), int(params.Position.Character))
	if word == "" {
		return nil, nil
	}

	// Builtins have no definition location in user code — return nothing.
	if _, isBuiltin := builtinDescriptions[word]; isBuiltin {
		return nil, nil
	}

	// Walk the AST and find the LetStatement that defines this name.
	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()

	for _, stmt := range program.Statements {
		letStmt, ok := stmt.(*ast.LetStatement)
		if !ok {
			continue
		}
		if letStmt.Name.Value != word {
			continue
		}

		// Found it. Return the position of the name token.
		// LSP is 0-based, our lexer is 1-based — subtract 1.
		nameToken := letStmt.Name.Token
		line := uint32(nameToken.Line - 1)
		col := uint32(nameToken.Col - 1)

		return protocol.Location{
			URI: protocol.DocumentUri(uri),
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: col + uint32(len(word))},
			},
		}, nil
	}

	// Name not found in this file.
	return nil, nil
}
