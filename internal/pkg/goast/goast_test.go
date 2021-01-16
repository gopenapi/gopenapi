package goast

import (
	"encoding/json"
	"github.com/zbysir/gopenapi/internal/pkg/gosrc"
	"testing"
)

func TestGoParse(t *testing.T) {
	goSrc, err := gosrc.NewGoSrcFromModFile("../../../")
	if err != nil {
		t.Fatal(err)
	}
	p := NewGoParse(goSrc)
	doc, exist, err := p.GetDoc("../../delivery/http/handler.PetHandler.FindPetByStatus")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s %t", doc, exist)
}

func TestParseStruct(t *testing.T) {
	goSrc, err := gosrc.NewGoSrcFromModFile("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	p := NewGoParse(goSrc)

	kc, _, err := p.GetStruct("../../model", "Pet")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(kc, " ", " ")
	t.Logf("%s", bs)
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
