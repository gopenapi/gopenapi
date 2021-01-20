package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/zbysir/gopenapi/internal/pkg/goast"
	"github.com/zbysir/gopenapi/internal/pkg/jsonordered"
	"github.com/zbysir/gopenapi/internal/pkg/log"
	"go/ast"
	"sort"
)

type ObjectSchema struct {
	Ref string `json:"$ref,omitempty"`

	Type string `json:"type"`

	Properties jsonordered.MapSlice `json:"properties"`
}

func (o *ObjectSchema) _schema() {}

func (a *ObjectSchema) GetType() string {
	return a.Type
}

// ObjectProp 对象成员
type ObjectProp struct {
	Schema
	Meta        jsonordered.MapSlice `json:"meta,omitempty"`
	Description string               `json:"description,omitempty"`
	Tag         map[string]string    `json:"tag,omitempty"`
	Example     interface{}          `json:"example,omitempty"`
}

// 对于嵌套了Interface的结构体, json不支持嵌入式序列化, 故出此下策.
type ObjectPropObj struct {
	*ObjectSchema
	Meta        jsonordered.MapSlice `json:"meta,omitempty"`
	Description string               `json:"description,omitempty"`
	Tag         map[string]string    `json:"tag,omitempty"`
	Example     interface{}          `json:"example,omitempty"`
}

type ObjectPropArray struct {
	*ArraySchema
	Meta        jsonordered.MapSlice `json:"meta,omitempty"`
	Description string               `json:"description,omitempty"`
	Tag         map[string]string    `json:"tag,omitempty"`
	Example     interface{}          `json:"example,omitempty"`
}

type ObjectPropIdent struct {
	*IdentSchema
	Meta        jsonordered.MapSlice `json:"meta,omitempty"`
	Description string               `json:"description,omitempty"`
	Tag         map[string]string    `json:"tag,omitempty"`
	Example     interface{}          `json:"example,omitempty"`
}

type ObjectPropNil struct {
	Meta        jsonordered.MapSlice `json:"meta,omitempty"`
	Description string               `json:"description,omitempty"`
	Tag         map[string]string    `json:"tag,omitempty"`
	Example     interface{}          `json:"example,omitempty"`
}

func (o ObjectProp) MarshalJSON() ([]byte, error) {
	switch s := o.Schema.(type) {
	case *ObjectSchema:
		return json.Marshal(ObjectPropObj{
			ObjectSchema: s,
			Meta:         o.Meta,
			Description:  o.Description,
			Tag:          o.Tag,
			Example:      o.Example,
		})
	case *ArraySchema:
		return json.Marshal(ObjectPropArray{
			ArraySchema: s,
			Meta:        o.Meta,
			Description: o.Description,
			Tag:         o.Tag,
			Example:     o.Example,
		})
	case *IdentSchema:
		return json.Marshal(ObjectPropIdent{
			IdentSchema: s,
			Meta:        o.Meta,
			Description: o.Description,
			Tag:         o.Tag,
			Example:     o.Example,
		})
	case *NilSchema:
		return json.Marshal(ObjectPropNil{
			Meta:        o.Meta,
			Description: o.Description,
			Tag:         o.Tag,
			Example:     o.Example,
		})
	default:
		panic(fmt.Sprintf("uncase Schema Type in Marshal %T", o.Schema))
	}

	return nil, nil
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

type NilSchema struct {
}

func (n NilSchema) _schema() {
}

func (n NilSchema) GetType() string {
	return "nil"
}

func (a *ArraySchema) _schema() {}

// goAstToSchema 将goAst转为Schema
//
//   expr参数是goAst
//   exprInFile 是这个expr在哪一个文件中(必须是相对路径, 如github.com/zbysir/gopenapi/internal/model/pet.go), 这是为了识别到这个文件引入了哪些包.
func (o *OpenApi) goAstToSchema(expr ast.Expr, exprInFile string) (Schema, error) {
	switch s := expr.(type) {
	case *ast.ArrayType:
		schema, err := o.goAstToSchema(s.Elt, exprInFile)
		if err != nil {
			return nil, err
		}
		return &ArraySchema{
			Type:  "array",
			Items: schema,
		}, nil

	case *ast.Ident:
		// 标识
		// 如果是基础类型, 则返回, 否则还需要继续递归.
		if is, t := IsBaseType(s.Name); is {
			return &IdentSchema{
				Type: t,
				Enum: nil,
			}, nil
		}
		// 获取当前包下的结构体
		def, exist, err := o.goparse.GetDef(o.goparse.GetPkgFile(exprInFile), s.Name)
		if err != nil {
			return nil, err
		}
		if !exist {
			log.Warningf("can't found Type: %s", s.Name)
			return &NilSchema{}, nil
		}

		schema, err := o.goAstToSchema(def.Type, def.File)
		if err != nil {
			return nil, err
		}

		// 如果是基础类型, 则需要获取枚举值
		if id, ok := schema.(*IdentSchema); ok {
			// 查找Enum
			enum, err := o.goparse.GetEnum(o.goparse.GetPkgFile(exprInFile), s.Name)
			if err != nil {
				return nil, err
			}

			_, defValue := enum.FirstValue()

			id.Enum = enum.Values
			id.Default = defValue
		}

		return schema, err
	case *ast.SelectorExpr:
		// for model.T syntax
		pkgName := s.X.(*ast.Ident).Name

		pkgs, err := o.goparse.GetFileImportPkg(exprInFile)
		if err != nil {
			return nil, err
		}

		if pkg, ispkg := pkgs[pkgName]; ispkg {
			str, exist, err := PkgGetter{
				goparse: o.goparse,
				pkg:     pkg,
			}.GetStruct(s.Sel.Name)
			if err != nil || !exist {
				return &NilSchema{}, err
			}

			schema, err := o.goAstToSchema(str.Type, str.File)
			return schema, err
		}

		return &NilSchema{}, nil
	case *ast.StructType:
		var props jsonordered.MapSlice

		for _, f := range s.Fields.List {
			p, err := o.goAstToSchema(f.Type, exprInFile)
			if err != nil {
				return nil, err
			}

			gd, err := o.parseGoDoc(f.Doc.Text(), exprInFile)
			if err != nil {
				return nil, err
			}

			props = append(props, jsonordered.MapItem{
				Key: f.Names[0].Name,
				Val: ObjectProp{
					Schema:      p,
					Meta:        gd.Meta,
					Description: gd.FullDoc,
					Tag:         encodeTag(f.Tag),
				},
			})
		}

		return &ObjectSchema{
			Type:       "object",
			Properties: props,
		}, nil
	default:
		panic(fmt.Sprintf("uncased goAstToSchema type: %T, %+v", s, s))
	}

	return &NilSchema{}, nil
}

type GoExprWithPath struct {
	goparse *goast.GoParse
	expr    ast.Expr
	path    string
	name    string
}

// 如果类型是 结构体, 则还需要查询到子方法, 或者子成员
func (g *GoExprWithPath) GetMember(k string) interface{} {
	str, ok := g.expr.(*ast.StructType)
	if !ok {
		return nil
	}

	// 返回子成员
	for _, field := range str.Fields.List {
		if k == field.Names[0].Name {
			return &GoExprWithPath{
				goparse: g.goparse,
				expr:    field.Type,
				path:    g.path,
				name:    k,
			}
		}
	}

	// 查找方法
	funcs, err := g.goparse.GetStructFunc(g.goparse.GetPkgFile(g.path), g.name)
	if err != nil {
		return nil
	}

	//
	log.Infof("funcs %+v", funcs)

	fun, exist := funcs[k]
	if !exist {
		return nil
	}

	return &GoExprWithPath{
		goparse: g.goparse,
		expr:    fun.Type,
		path:    g.path,
		name:    k,
	}
}

// 把任何格式的数据都转成Schema
func (o *OpenApi) anyToSchema(i interface{}) (Schema, error) {
	switch s := i.(type) {
	case *goast.Def:
		// just for test
		return o.goAstToSchema(s.Type, s.File)
	case *GoExprWithPath:
		return o.goAstToSchema(s.expr, s.path)
	case []interface{}:
		if len(s) == 0 {
			item, err := o.anyToSchema(nil)
			if err != nil {
				return nil, err
			}
			return &ArraySchema{
				Type:  "array",
				Items: item,
			}, nil
		}
		item, err := o.anyToSchema(s[0])
		if err != nil {
			return nil, err
		}
		return &ArraySchema{
			Type:  "array",
			Items: item,
		}, nil
	case map[string]interface{}:
		var keys []string
		for k := range s {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var props jsonordered.MapSlice
		for _, key := range keys {
			p, err := o.anyToSchema(s[key])
			if err != nil {
				return nil, err
			}
			props = append(props, jsonordered.MapItem{
				Key: key,
				Val: ObjectProp{
					Schema:      p,
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
		}, nil
	case string:
		return &IdentSchema{
			Type: "string",
			Enum: nil,
		}, nil
	case int64, int:
		return &IdentSchema{
			Type: "int",
			Enum: nil,
		}, nil
	case nil:
		return &ObjectSchema{
			Type:       "null",
			Properties: nil,
		}, nil
	default:
		panic(fmt.Sprintf("uncased type2Schema type: %T, %+v", s, s))
	}

}
