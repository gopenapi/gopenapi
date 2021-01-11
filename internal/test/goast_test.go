package test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"testing"
)

func TestGoAst(t *testing.T) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, "../delivery/http/handler", nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	// 扫描所有对象和他们的注释
	kc := map[string]*ast.CommentGroup{}
	for _, pkg := range pkgs {
		for _, v := range pkg.Files {
			for _, d := range v.Decls {
				switch d := d.(type) {
				case *ast.GenDecl:
					genDeclDoc := d.Doc
					spDoc := genDeclDoc
					for _, sp := range d.Specs {
						switch sp := sp.(type) {
						case *ast.TypeSpec:
							if sp.Doc != nil {
								spDoc = sp.Doc
							}

							kc[sp.Name.Name] = spDoc
						case *ast.ValueSpec:
							if sp.Doc != nil {
								spDoc = sp.Doc
							}

							// a,b,c = 1,2,3
							for _, name := range sp.Names {
								kc[name.Name] = spDoc
							}
						}
					}
				case *ast.FuncDecl:
					// 方法名
					key := d.Name.Name

					// 接受者
					for _, r := range d.Recv.List {
						switch r := r.Type.(type) {
						case *ast.StarExpr:
							key = r.X.(*ast.Ident).Name + "." + key
						}
					}
					kc[key] = d.Doc
				default:
					log.Panicf("uncased Decl type: %T", d)
				}
			}
		}
	}

	for k, c := range kc {
		t.Logf("%+v：%+v", k, c.Text())

		for _, l := range c.List {
			t.Logf("--- %+v：%+v", k, l.Text)
		}
	}
}
