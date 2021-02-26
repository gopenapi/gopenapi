package openapi

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

func TestRunJsExpress(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.conf.js")
	if err != nil {
		t.Fatal(err)
	}
	v, err := openAPi.runJsExpress("[...params(model.FindPetByStatusParams), {name: 'status', required: true}]",
		"github.com/gopenapi/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(v, " ", " ")
	t.Logf("%s", bs)
}

// 入口
func TestCompleteOpenapi(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.conf.js")
	if err != nil {
		t.Fatal(err)
	}

	bs, err := ioutil.ReadFile("../../../example/petstore/petstore_simp.yaml")
	if err != nil {
		t.Fatal(err)
	}
	dest, err := openAPi.CompleteYaml(string(bs))
	if err != nil {
		t.Fatal(err)
	}

	ioutil.WriteFile("TestConfig.yaml", []byte(dest), os.ModePerm)

	t.Logf("%s", dest)
}

// 测试 js表达式中, 结构体选择语法执行是否正确.
func TestToSchema(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.conf.js")
	if err != nil {
		t.Fatal(err)
	}
	v, err := openAPi.runJsExpress("model.Pet",
		"github.com/gopenapi/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(v, " ", " ")
	t.Logf("%s", bs)
}

func TestFullCommentMeta(t *testing.T) {
	// 读取openapi
	var kv []yaml.MapItem

	err := yaml.Unmarshal([]byte(`
$path:
  params: "js: [...params(model.FindPetByStatusParams), {name: 'status', required: true}]"
  js-resp: '{200: {desc: "成功", content: schema([model.Pet])}, 401: {desc: "没权限", content: schema({msg: "没权限"})}}'
`), &kv)
	if err != nil {
		t.Fatal(err)
	}
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}
	x, err := openAPi.fullCommentMeta(kv, "")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.MarshalIndent(x, "  ", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
	//t.Logf("%+v", x)
}

func TestGetGoDocForStruct(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}

	d, exist, err := openAPi.getGoStruct("github.com/gopenapi/gopenapi/internal/model.Pet", false)
	if err != nil {
		return
	}
	if !exist {
		t.Fatal("not exist")
	}
	bs, err := json.MarshalIndent(d, "  ", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}

func TestWorkSchemas(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}
	var kv []yaml.MapItem

	err = yaml.Unmarshal([]byte(`

components:
  schemas:
    Category:
      x-$schema: github.com/gopenapi/gopenapi/internal/model.Category
    Tag:
      x-$schema: github.com/gopenapi/gopenapi/internal/model.Tag
    Pet:
      x-$schema: github.com/gopenapi/gopenapi/internal/model.Pet
      required:
        - name
        - photoUrls

`), &kv)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = openAPi.walkSchemas(kv)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", openAPi.schemas)
}
