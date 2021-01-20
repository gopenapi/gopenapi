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
	openAPi, err := NewOpenApi("../../../go.mod")
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

// TODO CompleteOpenapi
func TestCompleteOpenapi(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	dest, err := openAPi.CompleteYaml(`
openapi: 3.0.1
info:
  title: Swagger Petstore
  description: 'This is a sample server Petstore server.  You can find out more about     Swagger
    at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).      For
    this sample, you can use the api key "special-key" to test the authorization     filters.'
  termsOfService: http://swagger.io/terms/
  contact:
    email: apiteam@swagger.io
  license:
    name: Apache 2.0
    url: http://www.apache.org/licenses/LICENSE-2.0.html
  version: 1.0.0
externalDocs:
  description: Find out more about Swagger
  url: http://swagger.io
servers:
  - url: https://petstore.swagger.io/v2
  - url: http://petstore.swagger.io/v2

tags:
  - name: pet
    description: Everything about your Pets
    externalDocs:
      description: Find out more
      url: http://swagger.io
  - name: store
    description: Access to Petstore orders
  - name: user
    description: Operations about user
    externalDocs:
      description: Find out more about our store
      url: http://swagger.io

paths:
  /pet/findByStatus:
    get:
      x-$path: github.com/zbysir/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus
      tags:
        - pet
      security:
        - petstore_auth:
            - write:pets
            - read:pets
  /pet/{petId}:
    get:
      x-$path: github.com/zbysir/gopenapi/internal/delivery/http/handler.PetHandler.GetPet
      tags:
        - pet
      operationId: getPetById
      security:
        - api_key: []
    post:
      tags:
        - pet
      summary: Updates a pet in the store with form data
      operationId: updatePetWithForm
      parameters:
        - name: petId
          in: path
          description: ID of pet that needs to be updated
          required: true
          schema:
            type: integer
            format: int64
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              properties:
                name:
                  type: string
                  description: Updated name of the pet
                status:
                  type: string
                  description: Updated status of the pet
      responses:
        405:
          description: Invalid input
          content: {}
      security:
        - petstore_auth:
            - write:pets
            - read:pets

components:
  schemas:
    Category:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
      xml:
        name: Category
    Tag:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
      xml:
        name: Tag
    Pet:
      x-$schema: github.com/zbysir/gopenapi/internal/model.Pet
      required:
        - name
        - photoUrls

  securitySchemes:
    petstore_auth:
      type: oauth2
      flows:
        implicit:
          authorizationUrl: http://petstore.swagger.io/oauth/dialog
          scopes:
            write:pets: modify pets in your account
            read:pets: read your pets
    api_key:
      type: apiKey
      name: api_key
      in: header
`)
	if err != nil {
		t.Fatal(err)
	}

	ioutil.WriteFile("TestConfig.yaml", []byte(dest), os.ModePerm)

	t.Logf("%s", dest)
}

func TestToSchema(t *testing.T) {
	openAPi, err := NewOpenApi("../../../go.mod")
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
	openAPi, err := NewOpenApi("../../../go.mod")
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
	openAPi, err := NewOpenApi("../../../go.mod")
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
	openAPi, err := NewOpenApi("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}

	d, exist, err := openAPi.GetGoDoc("github.com/zbysir/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus")
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
	openAPi, err := NewOpenApi("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}

	d, exist, err := openAPi.GetGoDoc("github.com/zbysir/gopenapi/internal/model.Pet")
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
