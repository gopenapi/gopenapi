package test

import (
	"encoding/json"
	yaml2 "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
	"testing"
)

// 测试输出有序的数据
func TestMap(t *testing.T) {
	var i []yaml.MapItem
	err := yaml.Unmarshal([]byte(`
$path:
  tags: [pet]
  params: | 
    js: {...{a: 1}, status: {required: true}}
  resp: 'js: {200: {desc: "成功", content: [model.Pet]}, 401: {desc: "没权限", content: {msg: "没权限"}}}'
`), &i)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := json.MarshalIndent(i, " ", " ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}

// 测试将yaml转为json
//   无序
func TestYToJson(t *testing.T) {
	j2, err := yaml2.YAMLToJSON([]byte(`
$path:
  params: {...model.FindPetByStatusParams, status: {required: true}, a: 1}
  resp: {"200": {z: 1, a: 2, desc: "成功", content: [model.Pet]}, 401: {desc: "没权限", content: {msg: "没权限"}}}
`))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", j2)
}
