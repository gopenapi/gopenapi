package openapi

import (
	"encoding/json"
	"github.com/dop251/goja"
	"go/ast"
	"go/parser"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

type Common struct {
	// Doc is Go comment without meta.
	Doc string
	// e.g.
	// $path
	//   params: {a: 1}
	Meta map[string]interface{}
}

func Meta() {

}

// 测试 扩展openapi语法
func TestY(t *testing.T) {
	var i []yaml.MapItem
	err := yaml.Unmarshal([]byte(`
$path:
  tags: [pet]
  params: | 
    js: [...model.FindPetByStatusParams, {name: status, required: true}]
  resp: 'js: {200: {desc: "成功", content: [model.Pet]}, 401: {desc: "没权限", content: {msg: "没权限"}}}'

`),
		&i)
	if err != nil {
		t.Fatal(err)
	}

	// 将go:语法转换为一个完整的json
	var r = full(i, "pet.go", map[string]struct{}{})

	bs, err := yaml.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", bs)

}

type OpenApi struct {
	paramsUseTag string
}

func TestXPathToOpenapi(t *testing.T) {
	x := XData{
		Summary:     "Finds Pets by status",
		Description: "Multiple status values can be provided with comma separated strings",
		Meta: map[string]interface{}{
			"parameters": ParamsList{
				{
					Name: "Status",
					Tag:  map[string]string{"json": "status"},
					Doc:  "Status values that need to be considered for filter",
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
	v, err := runJsExpress("[...model.FindPetByStatusParams, {name: 'status', required: true}]", "",
		map[string]struct{}{"parameters": {}})
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(v, " ", " ")
	t.Logf("%s", bs)
}

func TestType2Schema(t *testing.T) {
	e, err := parser.ParseExpr("x == struct{a int}{}")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", e.(*ast.BinaryExpr).Y.(*ast.CompositeLit).Type)
	s := anyToSchema(e.(*ast.BinaryExpr).Y.(*ast.CompositeLit).Type)

	bs, err := json.MarshalIndent(s, "  ", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}

func TestCompleteOpenapi(t *testing.T) {
	dest, err := CompleteOpenapi(`
paths:
  /pet/findByStatus:
    get:
      x-$path: internal/delivery/http/handler.PetHandler.FindPet

components:
  schemas:
    Pet:
      x-$schemas: internal/model.Pet
      required:
        - name
        - photoUrls
`)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", dest)
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

	x := fullCommentMeta(kv, "", map[string]struct{}{})
	bs, err := json.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
	//t.Logf("%+v", x)
}

func TestConfig(t *testing.T) {
	config := `
var config = {
	filter: function(key, value){
		if (key==='$path'){
			var responses = {}
		 	Object.keys(value.$path.resp).forEach(function (k){
				var v = value.$path.resp[k]
				responses[k] = {
				  description: v.desc,
				  content: {
				    'application/json':{
				      schema: v.content,
				    }
				  }
				}
			})
			return {
				parameters: value.$path.params.map(function (i){
					var x = i
					delete(x['_from'])
					if (x.tag) {
						if (x.tag.form){
							x.name = x.tag.form
						}
						delete(x['tag'])
					}
					if (x.doc){
						x.description = x.doc
						delete(x['doc'])
					}
					

					return x
				}),
				responses: responses
			}
		}
	}
}
`

	val := "var c = " + `{"$path":{"params":[{"_from":"go","name":"Status","tag":{"form":"status"},"doc":"Status values that need to be considered for filter\n$required: true\n","schema":{"type":"array","items":{"type":"PetStatus"}}},{"name":"status","required":true}],"resp":{"200":{"content":{"type":"array","items":{"type":"object","properties":{"Id":{"type":"int64","format":"int64","description":"Id is Pet ID\n","tag":{"json":"id"}},"Category":{"type":"Category","format":"Category","description":"Category Is pet category\n","tag":{"json":"category"}},"Name":{"type":"string","format":"string","description":"Id is Pet name\n","tag":{"json":"name"}},"Tags":{"type":"array","format":"","description":"Tag is Pet Tag\n","tag":{"json":"tags"}},"Status":{"type":"PetStatus","format":"PetStatus","description":"","tag":{"json":"status"}}}}},"desc":"成功"},"401":{"content":{"type":"object","properties":{"msg":{"type":"string","format":"string","description":"","example":"没权限"}}},"desc":"没权限"}}}}`
	jsCode := val + ";" + config + ";config.filter('$path', c)"

	gj := goja.New()
	v, err := gj.RunScript("js", jsCode)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := yaml.Marshal(v.Export())
	if err != nil {
		t.Fatal(err)
	}

	ioutil.WriteFile("TestConfig.yaml", bs, os.ModePerm)

	t.Logf("%s", bs)

}
