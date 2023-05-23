package main

import (
	"bike_race/core"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var CodeSmellsAnalyzer = &analysis.Analyzer{
	Name: "logging",
	Doc:  "Logging best practices",
	Run:  run,
}

func isIdent(node ast.Node, name string) bool {
	if ident, ok := node.(*ast.Ident); ok {
		return ident.Name == name
	}
	return false
}

func isLiteralSelector(node ast.Node, left string, right string) bool {
	if selector, ok := node.(*ast.SelectorExpr); ok {
		return isIdent(selector.X, left) && isIdent(selector.Sel, right)
	}
	return false
}

func checkCommandFunc(pass *analysis.Pass, node *ast.FuncDecl) {
	if node.Type.Results.NumFields() != 2 {
		pass.Reportf(node.Pos(), "command function must have 2 return value")
	} else {
		if !isIdent(node.Type.Results.List[0].Type, "int") {
			pass.Reportf(node.Pos(), "command function must return an int code as the first return value")
		}
		if !isIdent(node.Type.Results.List[1].Type, "error") {
			pass.Reportf(node.Pos(), "command function must return an error as the second return value")
		}
	}
	if userField := core.Find(node.Type.Params.List, func(field *ast.Field) bool {
		return isLiteralSelector(field.Type, "auth", "User")
	}); userField != nil {
		pass.Reportf(node.Pos(), "command function must not have an auth.User parameter")
	}
}

func checkQueryFunc(pass *analysis.Pass, node *ast.FuncDecl) {
	if node.Type.Results.NumFields() != 3 {
		pass.Reportf(node.Pos(), "query function must have 3 return values")
	} else {
		if !isIdent(node.Type.Results.List[1].Type, "int") {
			pass.Reportf(node.Pos(), "query function must return an int code as the second return value")
		}
		if !isIdent(node.Type.Results.List[2].Type, "error") {
			pass.Reportf(node.Pos(), "query function must return an error as the third return value")
		}
	}
}

func checkLogBeforePanic(pass *analysis.Pass, node *ast.BlockStmt) {
	for i, stmt := range node.List {
		if expr, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := expr.X.(*ast.CallExpr); ok {
				if i > 0 && isIdent(call.Fun, "panic") {
					if prev, ok := node.List[i-1].(*ast.ExprStmt); ok {
						if prevCall, ok := prev.X.(*ast.CallExpr); ok {
							if selector, ok := prevCall.Fun.(*ast.SelectorExpr); ok {
								if isIdent(selector.X, "slog") {
									pass.Reportf(call.Pos(), "no log before panic")
								} else if isIdent(selector.X, "logger") {
									pass.Reportf(call.Pos(), "no log before panic")
								}
							}
						}
					}
				}
			}
		}
	}
}

func visit(pass *analysis.Pass) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					if ident.Name == "log" {
						pass.Reportf(node.Pos(), "found old log usage")
					}
				}
			}
		case *ast.FuncDecl:
			if strings.HasSuffix(node.Name.Name, "Query") {
				checkQueryFunc(pass, node)
			}
			if strings.HasSuffix(node.Name.Name, "Command") {
				checkCommandFunc(pass, node)
			}
			if field := core.Find(node.Type.Params.List, func(field *ast.Field) bool {
				return isLiteralSelector(field.Type, "context", "Context")
			}); field != nil {
				if *field != node.Type.Params.List[0] {
					pass.Reportf(node.Pos(), "context.Context must be the first parameter")
				}
				if (*field).Names[0].Name != "ctx" {
					pass.Reportf(node.Pos(), "context.Context parameter must be named ctx")
				}
			}
		case *ast.BlockStmt:
			checkLogBeforePanic(pass, node)
		}
		return true
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, visit(pass))
	}
	return nil, nil
}
