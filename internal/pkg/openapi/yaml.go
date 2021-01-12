package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/zbysir/gopenapi/internal/pkg/goast"
	"github.com/zbysir/gopenapi/internal/pkg/js"
	"go/ast"
	"gopkg.in/yaml.v2"
	"log"
	"sort"
	"strings"
)

// 扩展openapi语法, 让其支持从go文件中读取注释信息

// ctx: 根据用户写的key, 只能匹配将结构体解析成什么格式
// 如 key中包含了parameters, 则将 struct解析成为 paramsList.
//    responses: 解析成 schema
func runJsExpress(code string, filename string, ctx map[string]struct{}) (interface{}, error) {
	//return nil,nil
	v, err := js.RunJs(code, func(name string, want string) interface{} {
		if name == "params" {
			return func(args ...interface{}) interface{} {
				stru := args[0]
				switch s := stru.(type) {
				case nil:
					return nil
				case ast.Expr:
					// case for:
					//   params(model.FindPetByStatusParams)
					return struct2ParamsList(s)
				}
				// 如果不是go的结构, 则原样返回
				//   params([{name: 'status'}])
				return stru
			}
		} else if name == "schema" {
			return func(args ...interface{}) interface{} {
				stru := args[0]
				return anyToSchema(stru)
			}
		}

		//res := func(i interface{}) interface{} {
		//	if _, ok := ctx["parameters"]; ok {
		//		switch i := i.(type) {
		//		case nil:
		//			return []interface{}{}
		//		case ast.Expr:
		//			// 从go解析出的结构
		//			//return []interface{}{}
		//			return struct2ParamsList(i)
		//		default:
		//			panic(fmt.Sprintf("uncase Type of parameters: %T", i))
		//		}
		//
		//		return i
		//	}
		//
		//	switch i := i.(type) {
		//	case ast.Expr:
		//		return type2Schema(i)
		//	default:
		//		return i
		//	}
		//}

		gp := goast.NewGoParse()

		// todo 根据 filename 获取import的pkg和别名

		goStruct, exist, err := gp.GetStruct("../../" + name)
		if err != nil {
			log.Printf("[err] %v", err)
			return nil
		}
		if !exist {
			return nil
		}

		return goStruct.Type
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

// 将struct解析成 openapi.parameters
// 返回的是[]ParamsItem.
func struct2ParamsList(s ast.Expr) []interface{} {
	var l []interface{}
	switch s := s.(type) {
	case *ast.StructType:
		for _, f := range s.Fields.List {
			l = append(l, ParamsItem{
				From: "go",
				Name: f.Names[0].Name,
				Tag:  encodeTag(f.Tag),
				Doc:  f.Doc.Text(),
				// TODO decode meta
				Meta: map[string]interface{}{},
				//Schema: nil,
				Schema: anyToSchema(f.Type),
			})
		}
	default:
		panic(fmt.Sprintf("uncased struct2ParamsList type: %T, %+v", s, s))
	}

	return l
}

// 把任何格式的数据都转成Schema
func anyToSchema(i interface{}) Schema {
	switch s := i.(type) {
	case *ast.ArrayType:
		return &ArraySchema{
			Type:  "array",
			Items: anyToSchema(s.Elt),
		}
	case *ast.Ident:
		return &IdentSchema{
			Type:    s.Name,
			Default: nil,
			Enum:    nil,
		}
	case *ast.StructType:
		var props JsonItems
		for _, f := range s.Fields.List {
			p := anyToSchema(f.Type)
			format := ""
			if x, ok := p.(interface{ Format() string }); ok {
				format = x.Format()
			}
			props = append(props, Item{
				Key: f.Names[0].Name,
				Val: ObjectProp{
					// todo 自动匹配ref, 或者最后遍历生成ref
					Ref:         "",
					Type:        p.GetType(),
					Format:      format,
					Meta:        nil,
					Tag:         encodeTag(f.Tag),
					Description: f.Doc.Text(),
				},
			})
		}

		return &ObjectSchema{
			Type:       "object",
			Properties: props,
		}
	case []interface{}:
		if len(s) == 0 {
			return &ArraySchema{
				Type:  "array",
				Items: anyToSchema(nil),
			}
		}
		return &ArraySchema{
			Type:  "array",
			Items: anyToSchema(s[0]),
		}
	case map[string]interface{}:
		var keys []string
		for k := range s {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var props JsonItems
		for _, key := range keys {
			p := anyToSchema(s[key])
			format := ""
			if x, ok := p.(interface{ Format() string }); ok {
				format = x.Format()
			}
			props = append(props, Item{
				Key: key,
				Val: ObjectProp{
					Ref:         "",
					Type:        p.GetType(),
					Format:      format,
					Meta:        nil,
					Description: "",
					Tag:         nil,
					Example:     s[key],
				},
			})
		}

		return &ObjectSchema{
			Type:       "object",
			Properties: props,
		}
	case string:
		return &IdentSchema{
			Type:    "string",
			Default: "",
			Enum:    nil,
		}
	case int64, int:
		return &IdentSchema{
			Type:    "int",
			Default: 0,
			Enum:    nil,
		}
	case nil:
		return &ObjectSchema{
			Type:       "object",
			Properties: nil,
		}
	default:
		panic(fmt.Sprintf("uncased type2Schema type: %T, %+v", s, s))
	}
}

// fullCommentMeta 处理在注释中的meta
// 如下:
// $path
//   parameters: "js: [...model.FindPetByStatusParams, {name: 'status', required: true}]"
//   resp: 'js: {200: {desc: "成功", content: [model.Pet]}, 401: {desc: "没权限", content: {msg: "没权限"}}}'
// filename 指定当前注释在哪一个文件中, 会根据文件中import的pkg获取.
// 返回结构体给最后组装yaml使用
func fullCommentMeta(i []yaml.MapItem, filename string, key map[string]struct{}) JsonItems {
	r := full(i, filename, key)
	return yamlItemToJsonItem(r)
}

func yamlItemToJsonItem(i []yaml.MapItem) JsonItems {
	r := make(JsonItems, len(i))
	for i, item := range i {
		switch v := item.Value.(type) {
		case []yaml.MapItem:
			r[i] = Item{
				Key: item.Key.(string),
				Val: yamlItemToJsonItem(v),
			}
		default:
			r[i] = Item{
				Key: item.Key.(string),
				Val: v,
			}
		}
	}
	return r
}

func full(i []yaml.MapItem, filename string, key map[string]struct{}) []yaml.MapItem {
	var r []yaml.MapItem
	for _, item := range i {
		s := item.Key.(string)
		key[s] = struct{}{}

		if strings.HasPrefix(s, "js-") {
			vs, ok := item.Value.(string)
			if !ok {
				panic(fmt.Sprintf("parse yaml err: value that key(%s) with 'js-' prefix must be string type, but %T", s, item.Value))
			}
			item.Key = s[3:]
			item.Value = "js: " + vs
		}

		switch v := item.Value.(type) {
		case string:
			if strings.HasPrefix(v, "js: ") {
				// js表达式
				jsCode := strings.Trim(v[3:], " ")
				v, err := runJsExpress(jsCode, filename, key)
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
				Value: full(v, filename, key),
			})
		case []interface{}:
			r = append(r, yaml.MapItem{
				Key:   item.Key,
				Value: v,
			})
		default:
			panic(fmt.Sprintf("uncased Value type %T", v))

		}

		delete(key, s)
	}

	return r
}

type XData struct {
	Summary     string
	Description string

	Meta map[string]interface{}
}

// 从go结构体能读出的数据, 用于parameters
type ParamsItem struct {
	// From 表示此item来至那, 如 go
	From   string                 `json:"_from"`
	Name   string                 `json:"name"`
	Tag    map[string]string      `json:"tag"`
	Doc    string                 `json:"doc"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
	Schema Schema                 `json:"schema"`
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
		Value: t.Meta["in"],
	})
	r = append(r, yaml.MapItem{
		Key:   "description",
		Value: t.Doc,
	})
	r = append(r, yaml.MapItem{
		Key:   "required",
		Value: t.Meta["required"],
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

// openapi.params格式
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
	GetType() string
}

type ArraySchema struct {
	Ref   string `json:"$ref,omitempty"`
	Type  string `json:"type"`
	Items Schema `json:"items"`
}

func (a *ArraySchema) GetType() string {
	return a.Type
}

type Item struct {
	Key string
	Val interface{}
}

// JsonItems 用于实现对struct的json有序序列化
// 用于生成schemas.
type JsonItems []Item

func (j JsonItems) MarshalJSON() ([]byte, error) {
	var bs = []byte(`{}`)
	var err error
	for _, item := range j {
		itembs, err := json.Marshal(item.Val)
		if err != nil {
			return nil, err
		}

		//itembs:=[]byte(`{"a":1}`)
		bs, err = jsonparser.Set(bs, itembs, item.Key)
		if err != nil {
			return nil, err
		}
	}

	return bs, err
}

type ObjectSchema struct {
	Ref string `json:"$ref,omitempty"`

	Type string `json:"type"`

	Properties JsonItems `json:"properties"`
}

func (o *ObjectSchema) _schema() {}

func (a *ObjectSchema) GetType() string {
	return a.Type
}

type ObjectProp struct {
	// ref 是自动获取到的ref, 就算ref存在下方的其他字段也会存在, 所以你可以选择是否使用ref.
	Ref         string                 `json:"$ref,omitempty"`
	Type        string                 `json:"type"`
	Format      string                 `json:"format"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Description string                 `json:"description"`
	Tag         map[string]string      `json:"tag,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
}

// 基础类型, string / int
type IdentSchema struct {
	Ref string `json:"$ref,omitempty"`

	Type    string        `json:"type"`
	Default interface{}   `json:"default,omitempty"`
	Enum    []interface{} `json:"enum,omitempty"`
}

func (a *IdentSchema) GetType() string {
	return a.Type
}

func (s *IdentSchema) _schema() {}

func (s *IdentSchema) Format() string {
	return s.Type
}

func (a *ArraySchema) _schema() {}

// TODO 使用js脚本让用户可以自己写逻辑
// 将元数据转成openapi.params
func xDataToParams(t *XData, useTag string) []yaml.MapItem {
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

// 分割tag
func encodeTag(tag *ast.BasicLit) map[string]string {
	if tag == nil {
		return nil
	}

	tags := strings.Trim(tag.Value, "`")
	r := map[string]string{}

	for _, t := range strings.Split(tags, " ") {
		ss := strings.Split(t, ":")
		if len(ss) == 2 {
			r[ss[0]] = strings.Trim(ss[1], `"`)
		}
	}
	return r
}

// 完成openapi
func CompleteOpenapi(inYaml string) (dest string, err error) {
	// 读取openapi
	var kv []yaml.MapItem

	err = yaml.Unmarshal([]byte(inYaml), &kv)
	if err != nil {
		return "", err
	}

	newKv := full(kv, "", map[string]struct{}{})

	out, err := yaml.Marshal(newKv)
	if err != nil {
		return
	}

	dest = string(out)
	return
}
