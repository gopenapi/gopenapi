package goast

import (
	"fmt"
	"github.com/zbysir/gopenapi/internal/pkg/gosrc"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// GoParse Parse the go src to:
// - doc
// - struct
type GoParse struct {
	gosrc *gosrc.GoSrc
}

func NewGoParse(gosrc *gosrc.GoSrc) *GoParse {
	return &GoParse{gosrc: gosrc}
}

// GetDoc 获取目标元素的注释
// e.g.
//  pkgDir : 支持两种格式
//    go引入路径格式: github.com/zbysir/gopenapi/internal/delivery/http/handler
//    包绝对路径: Z:\golang\go_project\gopenapi\internal\delivery\http\handler
//  key: PetHandler.FindPetByStatus
// 废弃
func (g *GoParse) GetDoc(pkgDir string, key string) (doc string, exist bool, err error) {
	pkgDir, err = g.gosrc.MustGetAbsPath(pkgDir)
	if err != nil {
		return
	}

	kc, err := parseDirDoc(pkgDir)
	if err != nil {
		return
	}

	comment, ok := kc[key]
	if ok {
		return comment.Text(), true, nil
	}

	return
}

func (g *GoParse) getPkgInfo(pkgDir string) (pkg *Pkg, err error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, pkgDir, nil, parser.PackageClauseOnly)
	if err != nil {
		return
	}

	pkgName := ""
	for _, pkg := range pkgs {
		pkgName = pkg.Name
	}

	pkg = &Pkg{
		Dir:     pkgDir,
		PkgName: pkgName,
	}
	return
}

// GetStruct 获取struct结构
// pkgDir: ../delivery/http/handler, 或者 基于gomod的引入路径
// key:
//   PetHandler
//   or: PetHandler.FuncA
func (g *GoParse) GetStruct(pkgDir string, key string) (def *Def, exist bool, err error) {
	pkgDir, err = g.gosrc.MustGetAbsPath(pkgDir)
	if err != nil {
		return
	}

	pa := NewParseAll()
	err = pa.parse(pkgDir)
	if err != nil {
		return nil, false, err
	}

	fmt.Printf("def: %+v\n", pa.def)
	def, exist = pa.def[key]
	if !exist {
		return
	}

	return
}

type Pkg struct {
	// 包源代码地址
	Dir string
	// 包名
	PkgName string
}

// Pkgs
// key: 导入的名字, 如别名 和 ., value: 包信息
// 暂时不支持.的处理
type Pkgs map[string]*Pkg

// GetFileImportPkg 获取文件中所有导入的包.
// Tips: 目前只支持获取文件中导入的**本项目**的其他包.
// goFilePath: github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go
func (g *GoParse) GetFileImportPkg(filePath string) (pkgs Pkgs, err error) {
	absPath, err := g.gosrc.MustGetAbsPath(filePath)
	if err != nil {
		return
	}

	// todo get from cache if parsed dir before.
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, absPath, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return
	}
	pkgs = make(map[string]*Pkg)
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"'`)
		fmt.Printf("%s\n", importPath)

		p, exist, err := g.gosrc.GetAbsPath(importPath)
		if err != nil {
			return nil, err
		}
		if !exist {
			continue
		}

		// 通过解析包代码获取包名
		pkg, err := g.getPkgInfo(p)
		if err != nil {
			return nil, err
		}

		localName := pkg.PkgName
		if imp.Name != nil {
			localName = imp.Name.Name
		}

		pkgs[localName] = pkg
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
