package openapi

import (
	"encoding/json"
	"testing"
)

func TestParseComment(t *testing.T) {
	abc := `
Finds Pets by status

Multiple status values can be provided with comma separated strings

$:
  a: 
  js-b: |
    [1+1, ...params(model.FindPetByStatusParams)]
$c: 2
$js-d: '{a: 1, b: schema(model.Pet)}'
`

	x, err := ParseGoDoc(abc)
	if err != nil {
		t.Fatal(err)
		return
	}

	bs, _ := json.MarshalIndent(x.Meta, "  ", "  ")
	t.Logf("%s", bs)
}
