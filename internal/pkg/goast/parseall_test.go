package goast

import "testing"

func TestParseAll(t *testing.T) {
	pa := parseAll{
		def: map[string]*Def{},
		let: map[string]*Let{},
	}
	err := pa.parse("../../model")
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range pa.def {
		t.Logf("%s %+v", k, v)
	}

	for k, v := range pa.let {
		t.Logf("%s %+v", k, v)
	}
}
