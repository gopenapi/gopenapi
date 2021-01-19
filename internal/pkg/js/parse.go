package js

import (
	"fmt"
	"github.com/zbysir/goja-parser/ast"
	"github.com/zbysir/goja-parser/parser"
)

func ParseExpress(code string) (ast.Expression, error) {
	p, err := parser.ParseFile(nil, fmt.Sprintf("'%s'",code), fmt.Sprintf("(%s)", code), 0)
	if err != nil {
		return nil, err
	}

	return p.Body[0].(*ast.ExpressionStatement).Expression, nil
}
