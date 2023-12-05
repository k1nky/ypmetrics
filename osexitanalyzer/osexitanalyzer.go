// Статический анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
package osexitanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for direct call os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		// проверяем файлы только из пакета main
		if file.Name.String() != "main" {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.String() == "main" {
					// обходим функцию main
					ast.Inspect(x.Body, func(nn ast.Node) bool {
						switch xx := nn.(type) {
						case *ast.CallExpr:
							if s, ok := xx.Fun.(*ast.SelectorExpr); ok {
								if from, ok := s.X.(*ast.Ident); ok {
									if from.Name == "os" && s.Sel.Name == "Exit" {
										pass.Reportf(from.NamePos, "avoid calling os.Exit in the main function")
									}
								}
							}
						}
						return true
					})
					// функцию main проверили дальше обходить дерево не стоит
					return false
				}
			}
			return true
		})
	}

	return nil, nil
}
