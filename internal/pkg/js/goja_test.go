package js

import (
	"encoding/json"
	"github.com/dop251/goja"
	"testing"
)

func TestX(t *testing.T) {
	//g:=goja.New()
	p, err := goja.Parse("ab", "a={...{c:1}}")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(p.Body, " ", " ")
	t.Logf("%s", bs)
}

func TestY(t *testing.T) {
	p, err := goja.Compile("ab", "({...{c:1}})", false)
	if err != nil {
		t.Fatal(err)
	}

	g := goja.New()
	r, err := g.RunProgram(p)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(r.Export())
}
