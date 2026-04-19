package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/sachinbhankhar/noname/ast"
	"github.com/sachinbhankhar/noname/lexer"
	"github.com/sachinbhankhar/noname/parser"
	"github.com/sachinbhankhar/noname/token"
)

// documents holds the current content of every open file.
var documents = map[string]string{}

// debounce state: one timer per open file URI.
var (
	debounceTimers = map[string]*time.Timer{}
	debounceMu     sync.Mutex
)

func textDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	content := params.TextDocument.Text
	documents[uri] = content
	publishDiagnostics(ctx, uri, content)
	return nil
}

func textDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	if len(params.ContentChanges) > 0 {
		change := params.ContentChanges[len(params.ContentChanges)-1]
		if c, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
			documents[uri] = c.Text
			scheduleDiagnostics(ctx, uri, c.Text)
		}
	}
	return nil
}

// scheduleDiagnostics waits 300ms after the last change before running diagnostics.
// This prevents errors from appearing on half-typed tokens (e.g. the first / of //).
func scheduleDiagnostics(ctx *glsp.Context, uri, content string) {
	debounceMu.Lock()
	defer debounceMu.Unlock()

	if t, ok := debounceTimers[uri]; ok {
		t.Stop()
	}
	debounceTimers[uri] = time.AfterFunc(300*time.Millisecond, func() {
		publishDiagnostics(ctx, uri, content)
	})
}

func textDocumentDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	delete(documents, uri)
	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []protocol.Diagnostic{},
	})
	return nil
}

// undefinedRef holds an identifier that was used but never defined.
type undefinedRef struct {
	tok token.Token
	msg string
}

// publishDiagnostics parses the content, checks for undefined identifiers,
// and sends all errors to the editor.
func publishDiagnostics(ctx *glsp.Context, uri string, content string) {
	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()

	severity := protocol.DiagnosticSeverityError
	diagnostics := []protocol.Diagnostic{}

	// Phase 1: parse errors (syntax problems).
	for _, pe := range p.ParseErrors() {
		line := uint32(pe.Token.Line - 1)
		col := uint32(pe.Token.Col - 1)
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: col + uint32(len(pe.Token.Literal))},
			},
			Severity: &severity,
			Message:  pe.Message,
		})
	}

	// Phase 2: undefined identifier check (semantic problems).
	// Only run if there are no parse errors — a broken AST gives false positives.
	if len(p.ParseErrors()) == 0 {
		for _, ref := range findUndefined(program) {
			line := uint32(ref.tok.Line - 1)
			col := uint32(ref.tok.Col - 1)
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: line, Character: col},
					End:   protocol.Position{Line: line, Character: col + uint32(len(ref.tok.Literal))},
				},
				Severity: &severity,
				Message:  ref.msg,
			})
		}
	}

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         protocol.DocumentUri(uri),
		Diagnostics: diagnostics,
	})
}

// findUndefined walks the AST and returns every identifier reference
// that has no matching definition in scope.
func findUndefined(program *ast.Program) []undefinedRef {
	// Seed scope with all builtin names.
	scope := map[string]bool{}
	for name := range builtinDescriptions {
		scope[name] = true
	}

	var refs []undefinedRef
	for _, stmt := range program.Statements {
		checkStatement(stmt, scope, &refs)
	}
	return refs
}

func checkStatement(stmt ast.Statement, scope map[string]bool, refs *[]undefinedRef) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		// Check the value expression before adding the name to scope,
		// so `let x = x` is flagged.
		checkExpression(s.Value, scope, refs)
		scope[s.Name.Value] = true

	case *ast.ReturnStatement:
		checkExpression(s.ReturnValue, scope, refs)

	case *ast.ExpressionStatement:
		checkExpression(s.Expression, scope, refs)

	case *ast.BlockStatement:
		for _, inner := range s.Statements {
			checkStatement(inner, scope, refs)
		}
	}
}

func checkExpression(expr ast.Expression, scope map[string]bool, refs *[]undefinedRef) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		if !scope[e.Value] {
			*refs = append(*refs, undefinedRef{
				tok: e.Token,
				msg: fmt.Sprintf("identifier not found: %s", e.Value),
			})
		}

	case *ast.PrefixExpression:
		checkExpression(e.Right, scope, refs)

	case *ast.InfixExpression:
		checkExpression(e.Left, scope, refs)
		checkExpression(e.Right, scope, refs)

	case *ast.IfExpression:
		checkExpression(e.Condition, scope, refs)
		checkStatement(e.Consequence, scope, refs)
		if e.Alternative != nil {
			checkStatement(e.Alternative, scope, refs)
		}

	case *ast.FunctionLiteral:
		// Function body gets a new scope that includes the parameters.
		inner := copyScope(scope)
		for _, param := range e.Parameters {
			inner[param.Value] = true
		}
		checkStatement(e.Body, inner, refs)

	case *ast.CallExpression:
		checkExpression(e.Function, scope, refs)
		for _, arg := range e.Arguments {
			checkExpression(arg, scope, refs)
		}
	}
}

func copyScope(s map[string]bool) map[string]bool {
	c := make(map[string]bool, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}
