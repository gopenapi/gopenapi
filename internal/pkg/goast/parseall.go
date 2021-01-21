package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"sync"
)

// 存储所有类型定义和变量/常量
type parseAll struct {
	cache sync.Map
}

func NewParseAll() *parseAll {
	return &parseAll{
	}
}

// 所有的定义
//  方法
//  类型
type Def struct {
	Name string
	// Key 是 唯一标识. e.g. github.com/zbysir/gopenapi/internal/model.Tag
	Key  string
	Type ast.Expr `json:"-"`

	// 只有方法定义有这个值
	FuncRecv *ast.FieldList `json:"-"`
	// 定义在哪个文件
	File string
	Doc  *ast.CommentGroup
}

// 变量以及常量
type Let struct {
	Value interface{}
	Type  ast.Expr
	Name  string
	// 定义在哪个文件
	File string
	Doc  *ast.CommentGroup
}

type cacheStruct struct {
	defs map[string]*Def
	let  []*Let
}

// 参数
//  path: 包文件地址
//
// 返回:
//	defs: 所有的类型定义(包括方法)
//	let: 所有的变量/常量
func (p *parseAll) parse(path string) (defs map[string]*Def, let []*Let, err error) {
	v, ok := p.cache.Load(path)
	if ok {
		s := v.(*cacheStruct)
		return s.defs, s.let, nil
	}
	defer func() {
		if err != nil {
			p.cache.Store(path, &cacheStruct{
				defs: defs,
				let:  let,
			})
		}
	}()

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return
	}

	defs = map[string]*Def{}

	for _, pkg := range pkgs {
		for filePath, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					genDeclDoc := decl.Doc

					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						// 声明类型
						case *ast.TypeSpec:
							// 如果没有单个的doc, 则使用外部的
							if spec.Doc == nil {
								spec.Doc = genDeclDoc
							}

							defs[spec.Name.Name] = &Def{
								Name:     spec.Name.Name,
								Key:      path + "." + spec.Name.Name,
								Type:     spec.Type,
								FuncRecv: nil,
								File:     filePath,
								Doc:      spec.Doc,
							}
						case *ast.ValueSpec:
							for i, name := range spec.Names {
								var value interface{}
								if len(spec.Values) > i {
									value = expr2Interface(spec.Values[i])
								}
								let = append(let, &Let{
									Value: value,
									Type:  spec.Type,
									Name:  name.Name,
									Doc:   spec.Doc,
									File:  filePath,
								})
							}
						case *ast.ImportSpec:

						default:
							panic(fmt.Sprintf("uncased spec type %T", spec))
						}
					}
				case *ast.FuncDecl:
					defs[decl.Name.Name] = &Def{
						Name: decl.Name.Name,
						Type: decl.Type,
						// TODO add decl.Recv name to key
						// path + "." + {Recv.name} + decl.Name.Name
						Key:      path + "." + decl.Name.Name,
						Doc:      decl.Doc,
						FuncRecv: decl.Recv,
						File:     filePath,
					}
				default:
					panic(fmt.Sprintf("uncased decl type %T", decl))
				}
			}
		}
	}

	return
}

// 将表达转为基础的类型
// 只支持 基础 类型 (ast.BasicLit)
func expr2Interface(expr ast.Expr) interface{} {
	switch expr := expr.(type) {
	case *ast.BasicLit:
		switch expr.Kind {
		case token.INT:
			i, _ := strconv.ParseInt(expr.Value, 10, 64)
			return i
		case token.FLOAT:
			i, _ := strconv.ParseFloat(expr.Value, 64)
			return i
		case token.IMAG:
			// 复数, 暂不处理
			return nil
		case token.CHAR:
			return strings.Trim(expr.Value, "'")
		case token.STRING:
			return strings.Trim(expr.Value, `"`)
		}
		return expr.Value
	}
	return expr
}
