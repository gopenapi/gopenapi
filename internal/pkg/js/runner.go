package js

import (
	"fmt"
	"github.com/zbysir/goja-parser/ast"
	"github.com/zbysir/goja-parser/token"
)

// RunJs 运行一个js表达式, 返回值
func RunJs(js string, getter func(name string) interface{}) (interface{}, error) {
	express, err := ParseExpress(js)
	if err != nil {
		return nil, err
	}

	r := Runner{getter: getter}

	return r.run(express)
}

type Runner struct {
	// Getter 是当js运行时遇到变量时调用的方法, 返回变量值
	getter func(name string) interface{}
}

func interface2ObjKey(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func (r *Runner) run(expression ast.Expression) (interface{}, error) {
	switch e := expression.(type) {
	case *ast.ArrayLiteral:
		var list []interface{}
		for _, v := range e.Value {
			switch v := v.(type) {
			case *ast.SpreadElement:
				arg, err := r.run(v.Argument)
				if err != nil {
					return nil, err
				}

				for _, a := range arg.([]interface{}) {
					list = append(list, a)
				}
			default:
				i, err := r.run(v)
				if err != nil {
					return nil, err
				}
				list = append(list, i)
			}
		}
		return list, nil
	case *ast.ObjectLiteral:
		var obj = map[string]interface{}{}
		for _, p := range e.Value {
			switch p := p.(type) {
			case *ast.ObjectProperty:
				key, err := r.run(p.Key)
				if err != nil {
					return nil, err
				}

				val, err := r.run(p.Value)
				if err != nil {
					return nil, err
				}
				obj[interface2ObjKey(key)] = val
			case *ast.SpreadElement:
				arg, err := r.run(p.Argument)
				if err != nil {
					return nil, err
				}

				for k, v := range arg.(map[string]interface{}) {
					obj[k] = v
				}
			}
		}
		return obj, nil
	case *ast.BooleanLiteral:
		return e.Value, nil
	case *ast.NumberLiteral:
		return e.Value, nil
	case *ast.StringLiteral:
		return e.Value.String(), nil
	case *ast.Identifier:
		return r.getter(e.Name.String()), nil
	case *ast.BinaryExpression:

		left, err := r.run(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := r.run(e.Right)
		if err != nil {
			return nil, err
		}

		switch e.Operator {
		case token.PLUS:
			return add(left, right), nil
		}
	case *ast.DotExpression:
		key := r.dotExpressionToString(e)
		return r.getter(key), nil
	case *ast.CallExpression:
		// 调用方法
		funci, err := r.run(e.Callee)
		if err != nil {
			return nil, err
		}

		fun := funci.(func(arg ...interface{}) interface{})
		args := make([]interface{}, len(e.ArgumentList))

		for i, a := range e.ArgumentList {
			arg, err := r.run(a)
			if err != nil {
				return nil, err
			}
			args[i] = arg
		}
		return fun(args...), nil
	default:
		panic(fmt.Sprintf("uncased expression Type :%T, %+v", e, e))
	}
	return nil, nil
}

func (r *Runner) dotExpressionToString(de *ast.DotExpression) string {
	key := de.Identifier.Name.String()
	switch l := de.Left.(type) {
	case *ast.Identifier:
		key = l.Name.String() + "." + key
	default:
		panic(fmt.Sprintf("uncased DotExpression.Left Type :%T", l))
	}
	return key
}

func add(a, b interface{}) interface{} {
	switch a := a.(type) {
	case float64:
		switch b := b.(type) {
		case float64:
			return a + b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a + b
		}

	}

	return fmt.Sprintf("%v%v", a, b)
}
