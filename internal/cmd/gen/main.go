package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	src := os.Args[1]
	pkgName := os.Args[2]
	varName := os.Args[3]

	bs, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}

	bs = bytes.ReplaceAll(bs, []byte(`"`), []byte(`\"`))
	bs = bytes.ReplaceAll(bs, []byte("\r\n"), []byte(`\n`))

	ioutil.WriteFile("gen.go", []byte(fmt.Sprintf("package %s\n\nconst %s = \"%s\"\n\n", pkgName, varName, bs)), os.ModePerm)
}
