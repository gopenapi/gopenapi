package openapi

import (
	"github.com/zbysir/gopenapi/internal/pkg/jsonordered"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

// 处理 go 注释为元数据
// 这个值是js脚本的参数, 故需要正确的被json序列化.
type GoStruct struct {
	// FullDoc 是整个注释(除去变量部分)
	FullDoc string `json:"doc"`

	// FullDoc的第一句
	Summary string `json:"summary"`
	// FullDoc除了第一个的剩下注释
	Description string `json:"description"`

	// Meta 是变量部分, 应该使用json序列化后(给js脚本)使用
	Meta jsonordered.MapSlice `json:"meta"`

	Schema    Schema `json:"schema,omitempty"`
	XGoStruct bool   `json:"x-gostruct"`
}

// parseGoDoc 将注释转为 纯注释文本 和 支持json序列化的Meta.
// 支持的注释格式如下:
//
// Multiple status values can be provided with comma separated strings
//
// $:
//   js-params: "[...params(model.FindPetByStatusParams), {name: 'status', required: true}]"
//   js-resp: '{200: {desc: "成功", content: schema([model.Pet]}, 401: {desc: "没权限", content: schema({msg: "没权限"})}}'
//
func (o *OpenApi) parseGoDoc(doc string, filepath string) (*GoStruct, error) {
	// 逐行扫描
	// 获取doc或者meta
	lines := strings.Split(doc, "\n")
	startIndent := 0
	open := false
	openLine := 0
	var yamlPart []string
	var pureDoc strings.Builder
	for i, line := range lines {
		if !open {
			if valReg.MatchString(line) {
				startIndent = getPrefixCount(line, ' ')
				openLine = i
				open = true
			} else {
				pureDoc.WriteString(line)
				pureDoc.WriteString("\n")
			}
		} else {
			indent := getPrefixCount(line, ' ')
			if indent <= startIndent {
				yamlPart = append(yamlPart, strings.Join(lines[openLine:i], "\n"))

				if valReg.MatchString(line) {
					startIndent = getPrefixCount(line, ' ')
					openLine = i
				} else {
					open = false
				}
			}
		}
	}

	if open {
		yamlPart = append(yamlPart, strings.Join(lines[openLine:], "\n"))
	}

	// 处理yaml变量
	var yamlObj [][]yaml.MapItem
	for _, y := range yamlPart {
		r, err := o.parseYaml(y, filepath)
		if err != nil {
			return nil, err
		}

		yamlObj = append(yamlObj, r)
	}

	doc = strings.TrimSpace(pureDoc.String())

	// 取出第一句作为Summary
	docX := doc
	lines = strings.Split(docX, "\n\n")
	summary := lines[0]
	description := strings.Join(lines[1:], "\n\n")

	return &GoStruct{
		FullDoc:     doc,
		Summary:     summary,
		Description: description,
		Meta:        yamlItemToJsonItem(combinObj(yamlObj...)),
		Schema:      nil,
		XGoStruct:   true,
	}, nil
}

// 组合多个yaml对象
func combinObj(o ...[]yaml.MapItem) []yaml.MapItem {
	// 判断重复, 重复直接覆盖.
	//keyMap := map[string]int{}
	//
	//var r []yaml.MapItem
	//for _, item := range o {
	//	for _, item := range item {
	//		if _, ok := keyMap[item.Key.(string)]; ok {
	//			r[keyMap[item.Key.(string)]] = item
	//			keyMap[item.Key.(string)] = len(r) - 1
	//		} else {
	//			r = append(r, item)
	//			keyMap[item.Key.(string)] = len(r) - 1
	//		}
	//	}
	//}

	var r []yaml.MapItem
	for _, item := range o {
		r = append(r, item...)
	}
	return r
}

// parseYaml: 处理yaml中的js表达式
func (o *OpenApi) parseYaml(y string, filepath string) ([]yaml.MapItem, error) {
	var i []yaml.MapItem
	err := yaml.Unmarshal([]byte(y),
		&i)
	if err != nil {
		return nil, err
	}

	var allObj []yaml.MapItem
	// 删除 $符号
	// 删除顶级的$
	for _, item := range i {
		key := item.Key.(string)
		if key == "$" {
			if obj, ok := item.Value.([]yaml.MapItem); ok {
				allObj = append(allObj, obj...)
				continue
			}
		} else if strings.HasPrefix(key, "$") {
			item.Key = key[1:]
		}
		allObj = append(allObj, item)
	}

	// 将go:语法转换为一个完整的json
	fulled, err := o.fullCommentMeta(allObj, filepath)
	if err != nil {
		return nil, err
	}

	return fulled, nil
}

// match for
// - $:
// - $path:
// - $abc: 1
var valReg = regexp.MustCompile(`^ *\$.*:.*$`)

func getPrefixCount(s string, sub int32) int {
	var r int
	for i, v := range s {
		if v != sub {
			break
		}
		r = i
	}

	return r
}
