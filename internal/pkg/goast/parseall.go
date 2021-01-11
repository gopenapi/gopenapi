package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type parseAll struct {
	d map[string]ast.Spec
}

func (p *parseAll) parse(path string) (err error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
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

							p.d[spec.Name.Name] = spec
						default:
							panic(fmt.Sprintf("uncased %T", spec))
						}
					}
				}
			}
		}
	}

	return
}
