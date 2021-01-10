package js

import (
	"encoding/json"
	"github.com/zbysir/goja-parser/parser"
	"testing"
)

func TestName(t *testing.T) {
	p, err := parser.ParseFile(nil, " a=  1", "a = 1", 0)
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(p, "  ", "  ")
	t.Logf("%s", bs)
}
