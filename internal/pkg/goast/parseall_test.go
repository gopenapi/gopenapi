package goast

import "testing"

func TestParseAll(t *testing.T) {
	pa := NewParseAll()
	def, let, err := pa.parse("../../model")
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range def {
		t.Logf("%s %+v", k, v)
	}

	for _, v := range let {
		t.Logf("%+v", v)
	}
}
