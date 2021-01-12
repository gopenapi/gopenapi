package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)


type GoParse struct {
}

func NewGoParse() *GoParse {
	return &GoParse{}
}

// GetDoc 获取目标元素的注释
// key的格式是: 路径/包名/元素, e.g.:  ../delivery/http/handler.PetHandler.FindPetByStatus
func (g *GoParse) GetDoc(key string) (doc string, exist bool, err error) {
	path, member := splitPkgPath(key)

	// todo cache doc of the path

	kc, err := parseDirDoc(path)
	if err != nil {
		return
	}

	comment, ok := kc[member]
	if ok {
		return comment.Text(), true, nil
	}

	return
}

// GetStruct 获取struct结构
// key的格式是: 路径/包名/元素, e.g.:  ../delivery/http/handler.PetHandler
func (g *GoParse) GetStruct(key string) (def *Def, exist bool, err error) {
	path, member := splitPkgPath(key)

	pa := NewParseAll()
	err = pa.parse(path)
	if err != nil {
		return nil, false, err
	}

	def, exist = pa.def[member]
	if !exist {
		return
	}

	return
}

func parseDirDoc(path string) (kc map[string]*ast.CommentGroup, err error) {
	// todo cache the path
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return
	}

	// 扫描所有对象和他们的注释
	kc = map[string]*ast.CommentGroup{}
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
					panic(fmt.Sprintf("uncased Decl type: %T", d))
				}
			}
		}
	}

	return
}
