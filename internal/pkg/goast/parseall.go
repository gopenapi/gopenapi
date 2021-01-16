package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// 存储所有类型
//   TypeSpec: 类型申明
//   Func: 方法声明
type parseAll struct {
	// 所有的定义
	def map[string]*Def
	// 所有的变量/常量
	let map[string]*Let
}

func NewParseAll() *parseAll {
	return &parseAll{
		def: map[string]*Def{},
		let: map[string]*Let{},
	}
}

// 所有的定义
//  方法
//  类型
type Def struct {
	Name string
	Type ast.Expr `json:"-"`
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

func (p *parseAll) parse(path string) (err error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return
	}

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

							p.def[spec.Name.Name] = &Def{
								Name: spec.Name.Name,
								Type: spec.Type,
								Doc:  spec.Doc,
								File: filePath,
							}
						case *ast.ValueSpec:
							for i, name := range spec.Names {
								var value interface{}
								if len(spec.Values) > i {
									value = expr2Interface(spec.Values[i])
								}
								p.let[name.Name] = &Let{
									Value: value,
									Type:  spec.Type,
									Name:  name.Name,
									Doc:   spec.Doc,
									File:  filePath,
								}
							}
						case *ast.ImportSpec:

						default:
							panic(fmt.Sprintf("uncased spec type %T", spec))
						}
					}
				case *ast.FuncDecl:
					p.def[decl.Name.Name] = &Def{
						Name: decl.Name.Name,
						Type: decl.Type,
						Doc:  decl.Doc,
						File: filePath,
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
// 只支持 string / int / float 类型
func expr2Interface(expr ast.Expr) interface{} {
	// TODO expr2Interface
	return expr
}
