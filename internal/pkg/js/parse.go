package js

import (
	"fmt"
	"github.com/zbysir/goja-parser/ast"
	"github.com/zbysir/goja-parser/parser"
)

func parseExpress(code string) (ast.Expression, string, error) {
	source := fmt.Sprintf("(%s)", code)
	p, err := parser.ParseFile(nil, fmt.Sprintf("'%s'", code), source, 0)
	if err != nil {
		return nil, source, err
	}
	return p.Body[0].(*ast.ExpressionStatement).Expression, source, nil
}
