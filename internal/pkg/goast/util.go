package goast

import (
	"path"
	"strings"
)

// splitPkgPath 分割路径src 为path和包名
//  src: ../internal/pkg/goast.GoMeta
//  output: path: ../internal/pkg/goast, SelectorExpr: GoMeta
func splitPkgPath(src string) (pa, member string) {
	p1, filename := path.Split(src)
	ss := strings.Split(filename, ".")
	if len(ss) != 0 {
		p1 += ss[0]
		member = strings.Join(ss[1:], ".")
	}

	pa = p1
	return
}
