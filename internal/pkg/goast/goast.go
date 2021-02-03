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
	gosrc    *gosrc.GoSrc
	parseAll *parseAll
}

func NewGoParse(gosrc *gosrc.GoSrc) *GoParse {
	return &GoParse{
		gosrc:    gosrc,
		parseAll: NewParseAll(),
	}
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

// GetDef 获取结构体/方法定义
// pkgDir: 基于gomod的引入路径
// key:
//   PetHandler
//   or: PetHandler.FuncA
//   不支持查询结构体成员属性.
func (g *GoParse) GetDef(pkgDir string, key string) (def *Def, exist bool, err error) {
	pkgDir, err = g.gosrc.MustGetAbsPath(pkgDir)
	if err != nil {
		return
	}

	defs, _, exist, err := g.parseAll.parse(pkgDir)
	if err != nil {
		return nil, false, err
	}
	if !exist {
		return
	}

	// 如果key包含了., 说明还需要查询子方法
	kk := strings.Split(key, ".")

	def, exist = defs[kk[0]]
	if !exist {
		return
	}

	def.File, err = g.gosrc.GetPkgPath(def.File)
	def.Key, err = g.gosrc.GetPkgPath(def.Key)

	if err != nil {
		return nil, false, err
	}

	if len(kk) > 1 {
		funcs, err := g.GetStructFunc(pkgDir, kk[0])
		if err != nil {
			return nil, false, err
		}

		fun, exist := funcs[kk[1]]
		if !exist {
			return nil, false, nil
		}
		return fun, true, nil
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

	_, let, exist, err := g.parseAll.parse(pkgDir)
	if err != nil {
		return nil, err
	}
	if !exist {
		return
	}
	enum = &Enum{
		Type:   typ,
		Values: nil,
	}

	for _, l := range let {
		if id, ok := l.Type.(*ast.Ident); ok {
			if id.Name == typ {
				enum.Values = append(enum.Values, l.Value)
				enum.Keys = append(enum.Keys, l.Name)
			}
		}
	}

	return
}

// GetStructFunc 获取结构体上的func
func (g *GoParse) GetStructFunc(pkgDir string, typName string) (enum map[string]*Def, err error) {
	pkgDir, err = g.gosrc.MustGetAbsPath(pkgDir)
	if err != nil {
		return
	}

	defs, _, exist, err := g.parseAll.parse(pkgDir)
	if err != nil {
		return nil, err
	}
	if !exist {
		return
	}

	enum = map[string]*Def{}
	for _, d := range defs {
		switch d.Type.(type) {
		case *ast.FuncType:
			if len(d.FuncRecv.List) != 0 {
				expr := d.FuncRecv.List[0].Type

				recvName := ""
				switch expr := expr.(type) {
				case *ast.Ident:
					recvName = expr.Name
				case *ast.StarExpr:
					recvName = expr.X.(*ast.Ident).Name
				default:
					panic(fmt.Sprintf("uncased Type of FuncRecv: %T", expr))
				}
				if recvName == typName {
					enum[d.Name] = d
				}
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

		p, isInProject, err := g.gosrc.GetAbsPath(importPath)
		if err != nil {
			return nil, err
		}
		// 不处理非本项目的包
		if !isInProject {
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

// GetPkgFile 获取文件所在的pkg
// github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go 返回 github.com/zbysir/gopenapi/internal/delivery/http/handler
func (g *GoParse) GetPkgFile(filePath string) (pkg string) {
	x := strings.LastIndexByte(filePath, '/')
	if x == -1 {
		return filePath
	}

	return filePath[:x]
}
