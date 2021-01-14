package openapi

import (
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

// 处理 go 注释为元数据

type GoDoc struct {
	// Doc 是整个注释(处理变量部分)
	Doc string `json:"doc"`
	// Meta 是变量部分, 应该使用json序列化后(给js脚本)使用
	Meta JsonItems
}

func ParseGoDoc(doc string) (*GoDoc, error) {
	// 逐行扫描
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
	var yamlObj []JsonItems
	for _, y := range yamlPart {
		r, err := parseYaml(y)
		if err != nil {
			return nil, err
		}

		yamlObj = append(yamlObj, r)
	}

	return &GoDoc{
		Doc:  strings.TrimSpace(pureDoc.String()),
		Meta: combinObj(yamlObj...),
	}, nil
}

// 组合多个yaml对象
func combinObj(o ...JsonItems) JsonItems {
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

	var r JsonItems
	for _, item := range o {
		r = append(r, item...)
	}
	return r
}

func parseYaml(y string) (JsonItems, error) {
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
	var fulled = full(allObj, "pet.go", map[string]struct{}{})

	return yamlItemToJsonItem(fulled), nil
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
