package js

import (
	"encoding/json"
	"github.com/zbysir/goja-parser/parser"
	"testing"
)

func TestName(t *testing.T) {
	p, err := parser.ParseFile(nil, " a =  1", "a = 1", 0)
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(p.Body, "  ", "  ")
	t.Logf("%s", bs)
}

func TestParseExpress(t *testing.T) {
	e, _, err := parseExpress("[...{a: 1}, a[1]]")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(e, "  ", "  ")
	t.Logf("%s", bs)
}
