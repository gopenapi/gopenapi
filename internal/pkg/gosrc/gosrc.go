package gosrc

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

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
// path: github.com/zbysir/gopenapi/internal/model, returned: Z:\golang\go_project\gopenapi\internal\model
// path: github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go, returned: Z:\golang\go_project\gopenapi\internal\delivery\http\handler\pet.go
func (s *GoSrc) GetAbsPath(path string) (absDir string, exist bool, err error) {
	if !strings.HasPrefix(path, s.ModuleName) {
		return
	}

	absDir = filepath.Clean(strings.ReplaceAll(path, s.ModuleName, s.AbsModuleFileDir))
	exist = true
	return
}

func (s *GoSrc) MustGetAbsPath(path string) (absDir string, err error) {
	absDir, eixst, err := s.GetAbsPath(path)
	if err != nil {
		return
	}
	if !eixst {
		err = fmt.Errorf("can't resove path: '%s', please ensure it starts with '%s'", path, s.ModuleName)
		return
	}
	return
}
