package openapi

import (
	"encoding/json"
	"github.com/zbysir/gopenapi/internal/model"
	"github.com/zbysir/gopenapi/internal/pkg/jsonordered"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestXPathToOpenapi(t *testing.T) {
	x := XData{
		Summary:     "Finds Pets by status",
		Description: "Multiple status values can be provided with comma separated strings",
		Meta: map[string]interface{}{
			"parameters": ParamsList{
				{
					Name:        "Status",
					Tag:         map[string]string{"json": "status"},
					Description: "Status values that need to be considered for filter",
					Schema: &ArraySchema{
						Type: "array",
						Items: &IdentSchema{
							Type:    "string",
							Default: "available",
							Enum:    []interface{}{"available", "pending"},
						},
					},
				},
			},
		},
	}

	r := xDataToParams(&x, "json")

	bs, err := yaml.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", bs)
}

func TestRunJsExpress(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}
	v, err := openAPi.runJsExpress("[...params(model.FindPetByStatusParams), {name: 'status', required: true}]",
		"github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go")
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

func TestToSchema(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}
	v, err := openAPi.runJsExpress("schema(model.Pet)",
		"github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(v, " ", " ")
	t.Logf("%s", bs)
}

type SchemaTestStruct struct {
	A string `json:"a"`
	B int    `json:"b"`
	C interface{}
	D struct {
		D1 int64 `json:"d_1"`
	}
}

type SchemaTestStructNested struct {
	*SchemaTestStruct
	E int
}

// https://stackoverflow.com/questions/36866035/how-to-refer-to-enclosing-type-definition-recursively-in-openapi-swagger
type PageTree struct {
	Id        int64 `json:"id" xorm:"pk autoincr int(11)"`
	ProjectId int64 `json:"project_id" xorm:"int(11) unique index"`

	PageIdTree PageIdTree `json:"content_id_tree" xorm:"json"`

	*model.Categorys
}

type PageIdTree struct {
	Id       int64         `json:"id"`
	Children []*PageIdTree `json:"children"`
}

func TestAnyToSchema(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.conf.js")
	if err != nil {
		t.Fatal(err)
	}
	var x interface{}

	// expect: panic
	s, err := openAPi.anyToSchema(x)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", s)
}

func TestGoToSchema(t *testing.T) {
	// todo add more test cases
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.conf.js")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		k        string
		wantKeys []string
	}{
		//{k: "SchemaTestStruct", wantKeys: []string{"A", "B", "C", "D", "D1"}},
		//{k: "SchemaTestStructNested", wantKeys: []string{"A", "B", "C", "D", "D1", "E"}},
		{k: "PageTree", wantKeys: []string{"Id"}},
	}

	for _, c := range cases {
		t.Run(c.k, func(t *testing.T) {
			def, exist, err := openAPi.goparse.GetDef("./internal/pkg/openapi", c.k)
			if err != nil {
				t.Fatal(err)
				return
			}
			if !exist {
				t.Fatal("not exist")
				return
			}

			s, err := openAPi.anyToSchema(&GoExprWithPath{
				goparse: openAPi.goparse,
				expr:    def.Type,
				doc:     def.Doc,
				file:    def.File,
				name:    def.Name,
				key:     def.Key,
				noRef:   false,
			})
			if err != nil {
				t.Fatal(err)
				return
			}

			if objs, ok := s.(*ObjectSchema); ok {
				key := getJsonItemKey(objs.Properties)
				if strings.Join(c.wantKeys, ",") != strings.Join(key, ",") {
					t.Errorf("unexpect result: result: %v, expected: %v ", key, c.wantKeys)
					bs, _ := json.MarshalIndent(s, " ", " ")
					t.Errorf("please check it: %s", bs)
				}
			}
		})

	}

	t.Logf("%s", "ok")
}

func getJsonItemKey(o jsonordered.MapSlice) []string {
	var a []string
	for _, item := range o {
		a = append(a, item.Key)
		switch t := item.Val.(type) {
		case ObjectProp:
			if objs, ok := t.Schema.(*ObjectSchema); ok {
				a = append(a, getJsonItemKey(objs.Properties)...)
			}
		}
	}
	return a
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

func TestGetGoDocForFun(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}

	d, exist, err := openAPi.getGoStruct("github.com/zbysir/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus", nil)
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

func TestGetGoDocForStruct(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}

	d, exist, err := openAPi.getGoStruct("github.com/zbysir/gopenapi/internal/model.Pet", nil)
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
      x-$schema: github.com/zbysir/gopenapi/internal/model.Category
    Tag:
      x-$schema: github.com/zbysir/gopenapi/internal/model.Tag
    Pet:
      x-$schema: github.com/zbysir/gopenapi/internal/model.Pet
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
