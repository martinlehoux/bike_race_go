package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var LoggingAnalyzer = &analysis.Analyzer{
	Name: "logging",
	Doc:  "Logging best practices",
	Run:  run,
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
