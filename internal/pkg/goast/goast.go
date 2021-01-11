package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Type 是go所有类型的数据
// 包括了
// - struct
// - ...
type Type struct {
	Doc *ast.CommentGroup
	//Fields []*StructField
	// type 表示声明的是什么类型的东西
	//   - struct: 结构体
	//   - string: 字符串
	Type ast.Expr          `json:"-"`
	Name string            `json:"name"`
	Tag  map[string]string `json:"tag"`
}

// 标识
type Identer interface {
	_typer()
}

type StructIdent struct {
	Type   *ast.StructType
	Name   string
	Doc    *ast.CommentGroup
	Fields []StructField
}

func (s *StructIdent) _typer() {}

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

// GetType 获取目标类型的源信息
// 类型包括:
//   通过 type xxx xxx 语法声明的类型
// key的格式是: 路径/包名/元素, e.g.:  ../delivery/http/handler.PetHandler.FindPetByStatus
func (g *GoParse) GetSpec(key string) (s Specer, exist bool, err error) {
	path, member := splitPkgPath(key)
	kc, err := parseDirType(path)
	if err != nil {
		return
	}

	comment, ok := kc[member]
	if ok {
		return comment, true, nil
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

// 规则
type Specer interface {
	_specer()
}

type FuncSpec struct {
	Name string
	Doc  *ast.CommentGroup
}

type StructSpec struct {
	Name   string
	Doc    *ast.CommentGroup
	Fields []StructField
}

func (s *StructSpec) _specer() {}

type TypeSpec struct {
	Name string
	Type Specer
	Doc  *ast.CommentGroup
}

// LetSpec 包括了 const 与 var
type LetSpec struct {
	Type Specer
	Name string
	Doc  *ast.CommentGroup
}

// StructField 是 StructIdent 的 Fields
type StructField struct {
	Type Specer
	Name string
	Doc  *ast.CommentGroup
	Tag  map[string]string
}

// 分割tag
func encodeTag(tag string) map[string]string {
	tag = strings.Trim(tag, "`")
	r := map[string]string{}

	for _, t := range strings.Split(tag, " ") {
		ss := strings.Split(t, ":")
		if len(ss) == 2 {
			r[ss[0]] = strings.Trim(ss[1], `"`)
		}
	}
	return r
}
