package goast

import (
	"testing"
)

func TestGoParse(t *testing.T) {
	p := GoParse{}
	doc, exist, err := p.GetDoc("../../delivery/http/handler.PetHandler.FindPetByStatus")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s %t", doc, exist)
}

//func TestParseStruct(t *testing.T) {
//	kc, err := parseDirType("../../model")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	bs, _ := json.MarshalIndent(kc, " ", " ")
//	t.Logf("%s", bs)
//}
