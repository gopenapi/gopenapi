package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/zbysir/gopenapi/internal/pkg/goast"
	"github.com/zbysir/gopenapi/internal/pkg/gosrc"
	"github.com/zbysir/gopenapi/internal/pkg/js"
	"go/ast"
	"gopkg.in/yaml.v2"
	"sort"
	"strings"
)

type OpenApi struct {
	goparse *goast.GoParse
}

func NewOpenApi(gomodFile string) (*OpenApi, error) {
	goSrc, err := gosrc.NewGoSrcFromModFile(gomodFile)
	if err != nil {
		return nil, err
	}
	p := goast.NewGoParse(goSrc)
	return &OpenApi{goparse: p}, nil
}

// PkgGetter 实现了 GetMember 接口, 用来给js解析器执行 member 语法.
type PkgGetter struct {
	goparse *goast.GoParse
	pkg     *goast.Pkg
}

func (p *PkgGetter) GetMember(k string) interface{} {
	def, exist, err := p.goparse.GetStruct(p.pkg.Dir, k)
	if err != nil {
		fmt.Printf("[err] %v", err)
		return nil
	}
	if !exist {
		return nil
	}
	return def.Type
}

// 扩展openapi语法, 让其支持从go文件中读取注释信息
// params:
//  code: js表达式
//  filepath: 当前go文件的路径, 会根据当前文件引入的包识别js表达式使用的是哪个包.
// return:
//  可能是任何东西
func (o *OpenApi) runJsExpress(code string, filePath string) (interface{}, error) {
	//return nil,nil
	v, err := js.RunJs(code, func(name string) (interface{}, error) {
		// builtin func:
		// - params
		// - schema
		if name == "params" {
			return func(args ...interface{}) (interface{}, error) {
				stru := args[0]
				switch s := stru.(type) {
				case nil:
					return nil, nil
				case ast.Expr:
					// case for follow syntax:
					//   params(model.FindPetByStatusParams)
					return o.struct2ParamsList(s, filePath), nil
				}
				// 如果不是go的结构, 则原样返回
				//   params([{name: 'status'}])
				return stru, nil
			}, nil
		} else if name == "schema" {
			return func(args ...interface{}) (interface{}, error) {
				stru := args[0]
				return o.anyToSchema(stru, filePath), nil
			}, nil
		}

		// 获取当前文件所有引入的包
		pkgs, err := o.goparse.GetFileImportPkg(filePath)
		if err != nil {
			return nil, err
		}

		// 判断name是否是pkg别名
		if pkg, ispkg := pkgs[name]; ispkg {
			// 如果是pkg, 则进入解析go源码流程
			return &PkgGetter{
				goparse: o.goparse,
				pkg:     pkg,
			}, nil
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

// 将struct解析成 openapi.parameters
// 返回的是[]ParamsItem.
func (o *OpenApi) struct2ParamsList(s ast.Expr, filePath string) []interface{} {
	var l []interface{}
	switch s := s.(type) {
	case *ast.StructType:
		for _, f := range s.Fields.List {
			gd, err := o.parseGoDoc(f.Doc.Text(), filePath)
			if err != nil {
				fmt.Printf("[err] %v", err)
				return nil
			}
			l = append(l, ParamsItem{
				From:   "go",
				Name:   f.Names[0].Name,
				Tag:    encodeTag(f.Tag),
				Doc:    gd.Doc,
				Meta:   gd.Meta,
				Schema: o.anyToSchema(f.Type, filePath),
			})
		}
	default:
		panic(fmt.Sprintf("uncased struct2ParamsList type: %T, %+v", s, s))
	}

	return l
}

// 把任何格式的数据都转成Schema
func (o *OpenApi) anyToSchema(i interface{}, filePath string) Schema {
	switch s := i.(type) {
	case *ast.ArrayType:
		return &ArraySchema{
			Type:  "array",
			Items: o.anyToSchema(s.Elt, filePath),
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
			p := o.anyToSchema(f.Type, filePath)
			format := ""
			if x, ok := p.(interface{ Format() string }); ok {
				format = x.Format()
			}

			gd, err := o.parseGoDoc(f.Doc.Text(), filePath)
			if err != nil {
				fmt.Printf("[err] %v", err)
				return nil
			}
			props = append(props, Item{
				Key: f.Names[0].Name,
				Val: ObjectProp{
					// todo 自动匹配ref, 或者最后遍历生成ref
					Ref:         "",
					Type:        p.GetType(),
					Format:      format,
					Meta:        gd.Meta,
					Tag:         encodeTag(f.Tag),
					Description: gd.Doc,
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
				Items: o.anyToSchema(nil, filePath),
			}
		}
		return &ArraySchema{
			Type:  "array",
			Items: o.anyToSchema(s[0], filePath),
		}
	case map[string]interface{}:
		var keys []string
		for k := range s {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var props JsonItems
		for _, key := range keys {
			p := o.anyToSchema(s[key], filePath)
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

func (o *OpenApi) fullCommentMetaToJson(i []yaml.MapItem, filename string) JsonItems {
	r := o.fullCommentMeta(i, filename)
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

// fullCommentMetaToJson 处理在注释中的meta
// 如下:
// $path
//   parameters: "js: [...model.FindPetByStatusParams, {name: 'status', required: true}]"
//   resp: 'js: {200: {desc: "成功", content: [model.Pet]}, 401: {desc: "没权限", content: {msg: "没权限"}}}'
// filename 指定当前注释在哪一个文件中, 会根据文件中import的pkg获取.
// 返回结构体给最后组装yaml使用
func (o *OpenApi) fullCommentMeta(i []yaml.MapItem, filename string) []yaml.MapItem {
	var r []yaml.MapItem
	for _, item := range i {
		s := item.Key.(string)

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
				// 处理js 为yaml对象
				jsCode := strings.Trim(v[3:], " ")
				v, err := o.runJsExpress(jsCode, filename)
				if err != nil {
					panic(err)
				}

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
				Value: o.fullCommentMeta(v, filename),
			})
		case []interface{}:
			r = append(r, yaml.MapItem{
				Key:   item.Key,
				Value: v,
			})
		case int, int64, int32, uint, uint32, float32, float64, bool:
			r = append(r, item)
		case nil:
			r = append(r, item)
		default:
			panic(fmt.Sprintf("uncased Value type %T", v))
		}
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
	From   string            `json:"_from"`
	Name   string            `json:"name"`
	Tag    map[string]string `json:"tag"`
	Doc    string            `json:"doc"`
	Meta   JsonItems         `json:"meta,omitempty"`
	Schema Schema            `json:"schema"`
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

	in, exist := t.Meta.Get("in")
	if exist {
		r = append(r, yaml.MapItem{
			Key:   "in",
			Value: in,
		})
	}
	r = append(r, yaml.MapItem{
		Key:   "description",
		Value: t.Doc,
	})
	required, exist := t.Meta.Get("required")
	if exist {
		r = append(r, yaml.MapItem{
			Key:   "required",
			Value: required,
		})
	}
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

func (j JsonItems) Get(key string) (v interface{}, exist bool) {
	for _, item := range j {
		if item.Key == key {
			return item.Val, true
		}
	}

	return nil, false
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
	Ref         string            `json:"$ref,omitempty"`
	Type        string            `json:"type"`
	Format      string            `json:"format"`
	Meta        JsonItems         `json:"meta,omitempty"`
	Description string            `json:"description"`
	Tag         map[string]string `json:"tag,omitempty"`
	Example     interface{}       `json:"example,omitempty"`
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

// 完成openapi, 入口
// TODO
func CompleteOpenapi(inYaml string) (dest string, err error) {
	// 读取openapi
	var kv []yaml.MapItem

	err = yaml.Unmarshal([]byte(inYaml), &kv)
	if err != nil {
		return "", err
	}

	//newKv := fullCommentMeta(kv, "", map[string]struct{}{})

	//out, err := yaml.Marshal(newKv)
	//if err != nil {
	//	return
	//}
	//
	//dest = string(out)
	return
}
