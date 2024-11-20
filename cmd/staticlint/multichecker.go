package main

import (
	"encoding/json"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/ast/inspector"
	"honnef.co/go/tools/staticcheck"
	"os"
	"path/filepath"
)

// ExitAnalyzer - это анализатор, который находит вызовы os.Exit в функции main.
var ExitAnalyzer = &analysis.Analyzer{
	Name: "exitChecker",
	Doc:  "Checks for direct calls to os.Exit in the main package",
	Run:  RunExitChecker,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

// RunExitChecker - функция проверки, которая будет вызвана анализатором.
func RunExitChecker(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		if f.Name.Name != "main" {
			continue
		}

		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		// Ищем вызовы os.Exit
		inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
			call := n.(*ast.CallExpr)
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "Direct call to os.Exit in main package")
				}
			}
		})
	}

	return nil, nil
}

// Config — имя файла конфигурации, в котором перечислены подключаемые анализаторы библиотеки staticckeck
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string
}

// Запускать программу можно из дериктории staticlint командой multichecker ../../..., которая начнет проверку
// с корня проекта и проверит все файлы проекта.
func main() {
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	mychecks := []*analysis.Analyzer{
		ExitAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}

	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}
	// добавляем анализаторы из staticcheck, которые указаны в файле конфигурации
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	multichecker.Main(
		mychecks...,
	)
}
