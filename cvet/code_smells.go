package main

import (
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
			if strings.HasSuffix(node.Name.Name, "Command") {
				if node.Type.Results.NumFields() != 2 {
					pass.Reportf(node.Pos(), "command function must have 2 return value")
				} else {
					if !isIdent(node.Type.Results.List[0].Type, "int") {
						pass.Reportf(node.Pos(), "query function must return an int code as the first return value")
					}
					if !isIdent(node.Type.Results.List[1].Type, "error") {
						pass.Reportf(node.Pos(), "command function must return an error as the second return value")
					}
				}
			}
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
