package main

import (
	"go/ast"
	"strings"

	"github.com/samber/lo"
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

func isSelector(node ast.Node, left string, right string) bool {
	if selector, ok := node.(*ast.SelectorExpr); ok {
		return isIdent(selector.X, left) && isIdent(selector.Sel, right)
	}
	return false
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
			if isSelector(node.Fun, "slog", "With") || isSelector(node.Fun, "logger", "With") {
				for _, arg := range node.Args {
					if call, ok := arg.(*ast.CallExpr); ok {
						if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := selector.X.(*ast.Ident); ok {
								if ident.Name != "slog" {
									pass.Reportf(node.Pos(), "slog.With and logger.With must be called with a slog arg")
								}
							}
						}
					} else {
						pass.Reportf(node.Pos(), "slog.With and logger.With must be called with a slog arg")
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
			if field, ok := lo.Find(node.Type.Params.List, func(field *ast.Field) bool {
				return isSelector(field.Type, "context", "Context")
			}); ok {
				if field != node.Type.Params.List[0] {
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
