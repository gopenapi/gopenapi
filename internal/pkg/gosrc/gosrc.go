package gosrc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GoSrc 负责管理 包引入地址 和 实际文件地址 之间的关系
type GoSrc struct {
	// module名字
	ModuleName       string `json:"module_name"`
	AbsModuleFileDir string
}

// 只有当以ModuleName开头的引入路径, 才能被解析.

var moduleRe = regexp.MustCompile(`module[ \t]+([^\s]+)`)

func NewGoSrcFromModFile(modFile string) (*GoSrc, error) {
	body, err := ioutil.ReadFile(modFile)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(modFile)
	if err != nil {
		return nil, err
	}
	absDir := filepath.Dir(abs)

	match := moduleRe.FindSubmatch(body)
	if len(match) != 2 {
		err = fmt.Errorf("not found module declare in file '%s'", modFile)
		return nil, err
	}
	return &GoSrc{
		ModuleName:       string(match[1]),
		AbsModuleFileDir: absDir,
	}, nil
}

// 获取文件的绝对路径
// e.g.
//   path: github.com/zbysir/gopenapi/internal/model, returned: Z:\golang\go_project\gopenapi\internal\model
//   path: github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go, returned: Z:\golang\go_project\gopenapi\internal\delivery\http\handler\pet.go
//   path: ./internal/delivery/http/handler/pet.go, returned: Z:\golang\go_project\gopenapi\internal\delivery\http\handler\pet.go
// return:
//   isInProject: 是否是本项目的地址
func (s *GoSrc) GetAbsPath(path string) (absDir string, isInProject bool, err error) {
	if filepath.IsAbs(path) {
		return path, true, nil
	}
	fp, isInProject := s.FormatPath(path)
	if !isInProject {
		return
	}

	absDir = filepath.Clean(strings.ReplaceAll(fp, s.ModuleName, s.AbsModuleFileDir))
	isInProject = true

	return
}

// GetPkgPath 返回相对路径
func (s *GoSrc) GetPkgPath(absPath string) (pkgPath string, err error) {
	rel, err := filepath.Rel(s.AbsModuleFileDir, absPath)
	if err != nil {
		return
	}
	pkgPath = fmt.Sprintf("%s/%s", s.ModuleName, strings.ReplaceAll(rel, string(os.PathSeparator), "/"))
	return
}

func (s *GoSrc) MustGetAbsPath(path string) (absDir string, err error) {
	absDir, isInProject, err := s.GetAbsPath(path)
	if err != nil {
		return
	}
	if !isInProject {
		err = fmt.Errorf("can't resove path: '%s', please ensure it starts with '%s' or './'", path, s.ModuleName)
		return
	}
	return
}

// FormatPath 会格式化用户在yaml中写的路径为规范路径(即包含module名字的完整路径)
func (s *GoSrc) FormatPath(path string) (fp string, isInProject bool) {
	if strings.HasPrefix(path, s.ModuleName) {
		fp = path
		isInProject = true
	} else if strings.HasPrefix(path, "./") {
		fp = s.ModuleName + path[1:]
		isInProject = true
	}

	return
}
