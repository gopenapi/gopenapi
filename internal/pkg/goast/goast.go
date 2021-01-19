package goast

import (
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

	def, exist = pa.def[key]
	if !exist {
		return
	}

	def.File, err = g.gosrc.GetPkgPath(def.File)
	if err != nil {
		return nil, false, err
	}

	return
}

type Enum struct {
	Type   string        `json:"type"`
	Values []interface{} `json:"values"`
	Keys   []string      `json:"keys"`
}

// GetEnum 获取枚举值, 只支持基础类型.
func (g *GoParse) GetEnum(pkgDir string, typ string) (enum *Enum, err error) {
	pkgDir, err = g.gosrc.MustGetAbsPath(pkgDir)
	if err != nil {
		return
	}

	pa := NewParseAll()
	err = pa.parse(pkgDir)
	if err != nil {
		return nil, err
	}
	enum = &Enum{
		Type:   typ,
		Values: nil,
	}

	for _, l := range pa.let {
		if id, ok := l.Type.(*ast.Ident); ok {
			if id.Name == typ {
				enum.Values = append(enum.Values, l.Value)
				enum.Keys = append(enum.Keys, l.Name)
			}
		}
	}

	return
}

func (e *Enum) FirstValue() (string, interface{}) {
	if e == nil {
		return "", nil
	}
	if len(e.Values) == 0 {
		return "", nil
	}

	return e.Keys[0], e.Values[0]
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

		pkg.Dir, err = g.gosrc.GetPkgPath(pkg.Dir)
		if err != nil {
			return nil, err
		}

		pkgs[localName] = pkg
	}

	return
}

// github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go 返回 github.com/zbysir/gopenapi/internal/delivery/http/handler
func (g *GoParse) GetFileInPkg(filePath string) (pkg string) {
	x := strings.LastIndexByte(filePath, '/')
	if x == -1 {
		return filePath
	}

	return filePath[:x]
}
