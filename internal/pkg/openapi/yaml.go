package openapi

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

// 扩展openapi语法, 让其支持从go文件中读取注释信息

func runJsExpress(code string) (interface{}, error) {

	return nil, nil
}

func full(i []yaml.MapItem) []yaml.MapItem {
	var r []yaml.MapItem
	for _, item := range i {
		switch v := item.Value.(type) {
		case string:
			if strings.HasPrefix(v, "js:") {
				// js表达式
				jsCode := strings.Trim(v[2:], " ")
				v, err := runJsExpress(jsCode)
				if err != nil {
					panic(err)
				}
				// 处理js 为json对象
				r = append(r, yaml.MapItem{
					Key:   item.Key,
					Value: v,
				})
			} else {
				r = append(r, yaml.MapItem{
					Key:   item.Key,
					Value: v,
				})
			}
		case []yaml.MapItem:
			r = append(r, yaml.MapItem{
				Key:   item.Key,
				Value: full(v),
			})
		case []interface{}:
			r = append(r, yaml.MapItem{
				Key:   item.Key,
				Value: v,
			})
		default:
			panic(fmt.Sprintf("uncased Value type %T", v))

		}
	}

	return r
}

func goType2Array() {

}

type XPath struct {
	Summary     string
	Description string

	Meta map[string]interface{}
}

// 从go结构体能读出的数据, 用于parameters
type ParamsItem struct {
	Name        string            `json:"name"`
	Tag         map[string]string `json:"tag"`
	In          string            `json:"in"`
	Description string            `json:"description"`
	Required    bool              `json:"required"`
	Style       string            `json:"style"`
	Explode     bool              `json:"explode"`
	Schema      Schema            `json:"schema"`
}

func (t *ParamsItem) ToYaml(useTag string) []yaml.MapItem {
	var r []yaml.MapItem

	name := t.Name
	if useTag != "" {
		if t := t.Tag[useTag]; t != "" {
			name = strings.Split(t, ",")[0]
		}
	}
	r = append(r, yaml.MapItem{
		Key:   "name",
		Value: name,
	})

	r = append(r, yaml.MapItem{
		Key:   "in",
		Value: t.In,
	})
	r = append(r, yaml.MapItem{
		Key:   "description",
		Value: t.Description,
	})
	r = append(r, yaml.MapItem{
		Key:   "required",
		Value: t.Required,
	})
	//r = append(r, yaml.MapItem{
	//	Key:   "style",
	//	Value: t.style,
	//})
	r = append(r, yaml.MapItem{
		Key:   "schema",
		Value: t.Schema,
	})
	return r

}

type ParamsList []ParamsItem

func (p ParamsList) ToYaml(useTag string) interface{} {
	var r [][]yaml.MapItem
	for _, p := range p {
		r = append(r, p.ToYaml(useTag))
	}

	return r
}

type Schema interface {
	_schema()
}

type ArraySchema struct {
	Type  string `yaml:"type"`
	Items Schema `yaml:"items"`
}

type StringSchema struct {
	Type    string        `yaml:"type"`
	Default interface{}   `yaml:"default"`
	Enum    []interface{} `yaml:"enum"`
}

func (s *StringSchema) _schema() {}

func (a *ArraySchema) _schema() {}

// TODO 使用js脚本让用户可以自己写逻辑
func (t *XPath) ToYaml(useTag string) []yaml.MapItem {
	var r []yaml.MapItem

	r = append(r, yaml.MapItem{
		Key:   "tag",
		Value: t.Meta["tag"],
	})

	var summary interface{} = t.Summary
	if s := t.Meta["summary"]; s != nil {
		summary = s
	}
	r = append(r, yaml.MapItem{
		Key:   "summary",
		Value: summary,
	})

	var description interface{} = t.Description
	if s := t.Meta["description"]; s != nil {
		description = s
	}
	r = append(r, yaml.MapItem{
		Key:   "description",
		Value: description,
	})

	r = append(r, yaml.MapItem{
		Key:   "parameters",
		Value: t.Meta["parameters"].(ParamsList).ToYaml(useTag),
	})

	r = append(r, yaml.MapItem{
		Key:   "responses",
		Value: t.Meta["responses"],
	})

	return r
}
