package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

// go run ./internal/cmd/gen/main.go ./gopenapi.conf.js ./internal/cmd/gen.go cmd defaultConfig
func main() {
	src := os.Args[1]
	dest := os.Args[2]
	pkgName := os.Args[3]
	varName := os.Args[4]

	bs, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}

	bs = bytes.ReplaceAll(bs, []byte(`"`), []byte(`\"`))
	bs = bytes.ReplaceAll(bs, []byte("\r"), []byte(``))
	bs = bytes.ReplaceAll(bs, []byte("\n"), []byte(`\n`))

	ioutil.WriteFile(dest, []byte(fmt.Sprintf("package %s\n\nconst %s = \"%s\"\n\n", pkgName, varName, bs)), os.ModePerm)
}
