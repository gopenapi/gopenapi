package js

import (
	"encoding/json"
	"fmt"
	"github.com/zbysir/goja-parser/ast"
	"github.com/zbysir/goja-parser/token"
	"strconv"
)

// RunJs 运行一个js表达式, 返回值
func RunJs(js string, getter func(name string) (interface{}, error)) (interface{}, error) {
	express, source, err := parseExpress(js)
	if err != nil {
		return nil, err
	}

	r := Runner{getter: getter, source: source}

	return r.run(express)
}

type Runner struct {
	// Getter 是当js运行时遇到变量时调用的方法, 返回变量值
	getter func(name string) (interface{}, error)
	// strict 表示是否是严格模式, 严格模式下, 遇到的错都会被return, 非严格模式下, Runner会尽量的返回nil, 而不报错.
	strict bool
	source string
}

func interface2ObjKey(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

type MemberGetter interface {
	GetMember(k string) interface{}
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
				array, ok := arg.([]interface{})
				if !ok {
					err := fmt.Errorf("spread opeart only support on []interface, but: %T", arg)
					if r.strict {
						return nil, err
					}
					fmt.Printf("[err] %v\n", err)
					return nil, nil
				}

				for _, a := range array {
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
		return r.getter(e.Name.String())
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
			return interfaceAdd(left, right), nil
		}
	case *ast.DotExpression:
		left, err := r.run(e.Left)
		if err != nil {
			return nil, err
		}
		id := e.Identifier.Name.String()
		switch left := left.(type) {
		case map[string]interface{}:
			return left[id], nil
		case MemberGetter:
			return left.GetMember(id), nil
		default:
			// TODO 非严格模式
			// 思考: 非严格模式实际上可以不要
			pkgName:=r.source[e.Idx0()-1:e.Left.Idx1()-1]
			return nil, fmt.Errorf("can't read '%s' of type '%T' at script '%s', please make sure you import '%s' package", id, left, r.source[e.Idx0()-1:e.Idx1()-1], pkgName)
		}

	case *ast.CallExpression:
		// 调用方法
		funci, err := r.run(e.Callee)
		if err != nil {
			return nil, err
		}

		fun, ok := funci.(func(arg ...interface{}) (interface{}, error))
		if !ok {
			return nil, r.newError(int(e.Idx0()), int(e.Idx1()),
				fmt.Errorf("'%s' is not a function", r.expressionToString(e.Callee)))
		}
		args := make([]interface{}, len(e.ArgumentList))

		for i, a := range e.ArgumentList {
			arg, err := r.run(a)
			if err != nil {
				return nil, err
			}
			args[i] = arg
		}
		return fun(args...)
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

// 用于debug
func (r *Runner) expressionToString(de ast.Expression) string {
	switch l := de.(type) {
	case *ast.DotExpression:
		return r.expressionToString(l.Left) + "." + r.expressionToString(&l.Identifier)
	case *ast.Identifier:
		return l.Name.String()
	default:
		panic(fmt.Sprintf("uncased expression Type :%T", l))
	}
	return ""
}

func (r *Runner) newError(start, end int, err error) error {
	return fmt.Errorf("run '%s' err: %w", r.source[start-1:end-1], err)
}

func interfaceAdd(a, b interface{}) interface{} {
	an, ok := isNumber(a)
	if !ok {
		return interfaceToStr(a) + interfaceToStr(b)
	}
	bn, ok := isNumber(b)
	if !ok {
		return interfaceToStr(a) + interfaceToStr(b)
	}

	return an + bn
}

func isNumber(s interface{}) (d float64, is bool) {
	if s == nil {
		return 0, false
	}
	switch a := s.(type) {
	case int:
		return float64(a), true
	case int32:
		return float64(a), true
	case int64:
		return float64(a), true
	case float64:
		return a, true
	case float32:
		return float64(a), true
	default:
		return 0, false
	}
}

func interfaceToStr(s interface{}) (d string) {
	switch a := s.(type) {
	case string:
		d = a
	case int:
		d = strconv.FormatInt(int64(a), 10)
	case int32:
		d = strconv.FormatInt(int64(a), 10)
	case int64:
		d = strconv.FormatInt(a, 10)
	case float64:
		d = strconv.FormatFloat(a, 'f', -1, 64)
	default:
		bs, _ := json.Marshal(a)
		d = string(bs)
	}

	return
}
