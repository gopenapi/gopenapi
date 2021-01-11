package openapi

import (
	"gopkg.in/yaml.v2"
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
	var r = full(i)

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
	x := XPath{
		Summary:     "Finds Pets by status",
		Description: "Multiple status values can be provided with comma separated strings",
		Meta: map[string]interface{}{
			"parameters": ParamsList{
				{
					Name:        "Status",
					Tag:         map[string]string{"json": "status"},
					In:          "",
					Description: "Status values that need to be considered for filter",
					Required:    false,
					Style:       "",
					Explode:     false,
					Schema: &ArraySchema{
						Type: "array",
						Items: &StringSchema{
							Type:    "string",
							Default: "available",
							Enum:    []interface{}{"available", "pending"},
						},
					},
				},
			},
		},
	}

	r := x.ToYaml("json")

	bs, err := yaml.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", bs)
}
