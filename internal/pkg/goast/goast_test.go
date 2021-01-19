package goast

import (
	"encoding/json"
	"github.com/zbysir/gopenapi/internal/pkg/gosrc"
	"testing"
)

func TestParseStruct(t *testing.T) {
	goSrc, err := gosrc.NewGoSrcFromModFile("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	p := NewGoParse(goSrc)

	kc, exist, err := p.GetDef("github.com/zbysir/gopenapi/internal/delivery/http/handler", "PetHandler")
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

	pkgs, err := p.GetFileImportPkg("github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
	}
	bs, _ := json.MarshalIndent(pkgs, " ", " ")
	t.Logf("%s", bs)
}
