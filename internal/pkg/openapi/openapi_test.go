package openapi

import (
	"encoding/json"
	"github.com/dop251/goja"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

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

	bs, err := ioutil.ReadFile("../../../openapi/petstore_simp.yaml")
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

func TestAnyToSchema(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod", "../../../gopenapi.js")
	if err != nil {
		t.Fatal(err)
	}

	def, exist, err := openAPi.goparse.GetDef("github.com/zbysir/gopenapi/internal/model/modelt", "Pet")
	if err != nil {
		t.Fatal(err)
		return
	}
	if !exist {
		t.Fatal("not exist")
		return
	}

	s, err := openAPi.anyToSchema(def)
	if err != nil {
		t.Fatal(err)
		return
	}
	bs, _ := json.MarshalIndent(s, " ", " ")
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
	x := openAPi.fullCommentMetaToJson(kv, "")
	bs, err := json.MarshalIndent(x, "  ", "  ")
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

	val := "var c = " + `{"$path":{"params":[{"_from":"go","name":"Status","tag":{"form":"status"},"doc":"Status values that need to be considered for filter\n$required: true\n","schema":{"type":"array","items":{"type":"PetStatus"}}},{"name":"status","required":true}],"resp":{"200":{"content":{"type":"array","items":{"type":"object","properties":{"Id":{"type":"int64","format":"int64","description":"Id is Pet ID\n","tag":{"json":"id"}},"Category":{"type":"Category","format":"Category","description":"Category Is pet category\n","tag":{"json":"category"}},"PkgName":{"type":"string","format":"string","description":"Id is Pet name\n","tag":{"json":"name"}},"Tags":{"type":"array","format":"","description":"Tag is Pet Tag\n","tag":{"json":"tags"}},"Status":{"type":"PetStatus","format":"PetStatus","description":"","tag":{"json":"status"}}}}},"desc":"成功"},"401":{"content":{"type":"object","properties":{"msg":{"type":"string","format":"string","description":"","example":"没权限"}}},"desc":"没权限"}}}}`
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

func TestJsonObjectProp(t *testing.T) {
	x := ObjectProp{
		Meta:        nil,
		Description: "1111",
		Tag:         nil,
		Example:     nil,
		Schema: &ObjectSchema{
			Ref:        "x",
			Type:       "obj",
			Properties: nil,
		},
	}

	t.Logf("%#v", x)

	bs, err := json.MarshalIndent(&x, " ", " ")
	t.Logf("%s %v", bs, err)
}

func TestYaml(t *testing.T) {
	i := []yaml.MapItem{}

	// 不支持根不是对象的yaml
	err := yaml.Unmarshal([]byte(`
- 1
`), &i)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", i)
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
