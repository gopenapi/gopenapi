package goast

import (
	"encoding/json"
	"github.com/gopenapi/gopenapi/internal/pkg/gosrc"
	"testing"
)

func TestParseStruct(t *testing.T) {
	goSrc, err := gosrc.NewGoSrcFromModFile("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	p := NewGoParse(goSrc)

	kc, exist, err := p.GetDef("github.com/gopenapi/gopenapi/internal/delivery/http/handler", "PetHandler")
	if err != nil {
		t.Fatal(err)
	}
	if !exist {
		t.Fatal("not exist")
	}

	bs, _ := json.MarshalIndent(kc, " ", " ")
	t.Logf("%s %s", bs, kc.Doc.Text())
}

func TestGetFileImportPkg(t *testing.T) {
	goSrc, err := gosrc.NewGoSrcFromModFile("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	p := NewGoParse(goSrc)

	pkgs, err := p.GetFileImportedPkgs("github.com/gopenapi/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
	}
	bs, _ := json.MarshalIndent(pkgs, " ", " ")
	t.Logf("%s", bs)
}

func TestGetStructFunc(t *testing.T) {
	goSrc, err := gosrc.NewGoSrcFromModFile("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	p := NewGoParse(goSrc)

	pkgs, err := p.GetFuncOfStruct("github.com/gopenapi/gopenapi/internal/delivery/http/handler", "PetHandler")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.MarshalIndent(pkgs, " ", " ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}
