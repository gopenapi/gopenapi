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
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}
	x, err := openAPi.parseGoDoc(abc, "github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
		return
	}

	bs, _ := json.MarshalIndent(x.Meta, "  ", "  ")
	t.Logf("%s", bs)
}
