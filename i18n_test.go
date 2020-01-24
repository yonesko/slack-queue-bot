package main

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/i18n"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"
	"strings"
	"testing"
)

func TestAllLabelsAreUsedAndDefined(t *testing.T) {
	usedLabels := collectUsedLabels()
	if len(usedLabels) == 0 {
		t.Error("no labels are used, use some or skip this test")
	}
	for l, _ := range usedLabels {
		if val, ok := i18n.P.Get(l); !ok || len(strings.TrimSpace(val)) == 0 {
			t.Errorf("label %s is undefined", l)
		}
	}
	for _, l := range i18n.P.Keys() {
		if _, ok := usedLabels[l]; !ok {
			t.Errorf("label %s is unused\n", l)
		}
	}
}

func collectUsedLabels() map[string]struct{} {
	node, err := parser.ParseFile(token.NewFileSet(), "controller.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	usedLabels := map[string]struct{}{}
	if containsI18nPackage(node) {
		ast.Inspect(node, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fun, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if fun.Sel.Name == "MustGetString" {
				val, err := strconv.Unquote(callExpr.Args[0].(*ast.BasicLit).Value)
				if err != nil {
					panic(fmt.Errorf("strconv.Unquote %s", err))
				}
				usedLabels[val] = struct{}{}
			}
			return true
		})
	}
	return usedLabels
}

func containsI18nPackage(f *ast.File) bool {
	for _, imp := range f.Imports {
		if imp.Path.Value == "\"github.com/yonesko/slack-queue-bot/i18n\"" {
			return true
		}
	}
	return false

}
