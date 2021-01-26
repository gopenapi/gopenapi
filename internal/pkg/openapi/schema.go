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

type ObjectPropRef struct {
	Ref string `json:"$ref"`
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
	case *RefSchema:
		return json.Marshal(ObjectPropRef{
			Ref: s.Ref,
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

//type Expr struct {
//	// 表达式, 如 model.Pet
//	expr ast.Expr
//	// 当前表达式所在的文件, 当expr是 model.Pet 时, 会根据当前文件的import找到model包下的Pet结构体.
//	exprInFile string
//	key string
//}

func (expr *GoExprWithPath) Key() (string, error) {
	if expr.key != "" {
		return expr.key, nil
	}
	switch s := expr.expr.(type) {
	case *ast.ArrayType:
		// 不处理
		// 不支持数组生成ref.
		// 如需支持数组, 需要定义type: `type Items []Item`, 然后使用Items生成ref
		return "", nil
	case *ast.Ident:
		// 标识
		// 如果是基础类型, 则返回, 否则还需要继续递归.
		if is, _ := IsBaseType(s.Name); is {
			return s.Name, nil
		}

		// 当前包的结构体
		return expr.goparse.GetPkgFile(expr.file) + "." + s.Name, nil
	case *ast.SelectorExpr:
		// for model.T syntax
		pkgName := s.X.(*ast.Ident).Name

		pkgs, err := expr.goparse.GetFileImportPkg(expr.file)
		if err != nil {
			return "", err
		}

		if pkg, ispkg := pkgs[pkgName]; ispkg {
			return pkg.Dir + "." + s.Sel.Name, nil
		}

		return "", fmt.Errorf("can't found package: %s in file: %s", pkgName, expr.file)
	case *ast.StarExpr:
		expr.expr = s.X
		return expr.Key()
	default:
		panic(fmt.Sprintf("uncase Type of GetExprKey: %T", expr.expr))
	}

}

// goAstToSchema 将goAst转为Schema
//
//   expr参数是goAst
//   exprInFile 是这个expr在哪一个文件中(必须是相对路径, 如github.com/zbysir/gopenapi/internal/model/pet.go), 这是为了识别到这个文件引入了哪些包.
func (o *OpenApi) goAstToSchema(expr *GoExprWithPath) (Schema, error) {
	if !expr.noRef{
		exprKey, err := expr.Key()
		if err != nil {
			err = fmt.Errorf("call Expr.Key err: %w", err)
			return nil, err
		}
		if ref, ok := o.schemas[exprKey]; ok {
			return &RefSchema{
				Ref: "#/" + ref,
			}, nil
		}
	}

	switch s := expr.expr.(type) {
	case *ast.ArrayType:
		// TODO key
		schema, err := o.goAstToSchema(&GoExprWithPath{
			goparse: o.goparse,
			expr:    s.Elt,
			file:    expr.file,
			name:    "",
			key:     "",
		})
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
		def, exist, err := o.goparse.GetDef(o.goparse.GetPkgFile(expr.file), s.Name)
		// 获取当前包下的结构体
		if err != nil {
			return nil, err
		}
		if !exist {
			log.Warningf("can't found Type: %s", s.Name)
			return &NilSchema{}, nil
		}

		schema, err := o.goAstToSchema(&GoExprWithPath{
			goparse: o.goparse,
			expr:    def.Type,
			file:    def.File,
			name:    def.Name,
			key:     def.Key,
		})
		//schema, err := o.goAstToSchema(def.Type, def.File)
		if err != nil {
			return nil, err
		}

		// 如果是基础类型, 则需要获取枚举值
		if id, ok := schema.(*IdentSchema); ok {
			// 查找Enum
			enum, err := o.goparse.GetEnum(o.goparse.GetPkgFile(expr.file), s.Name)
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

		pkgs, err := o.goparse.GetFileImportPkg(expr.file)
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

			schema, err := o.goAstToSchema(&GoExprWithPath{
				goparse: o.goparse,
				expr:    str.Type,
				file:    str.File,
				name:    str.Name,
				key:     str.Key,
			})
			//schema, err := o.goAstToSchema(str.Type, str.File)
			return schema, err
		}

		return &NilSchema{}, nil
	case *ast.StructType:
		var props jsonordered.MapSlice

		for _, f := range s.Fields.List {
			name := f.Names[0].Name
			p, err := o.goAstToSchema(&GoExprWithPath{
				goparse: o.goparse,
				expr:    f.Type,
				file:    expr.file,
				name:    name,
				key:     "",
			})
			//p, err := o.goAstToSchema(f.Type, exprInFile)
			if err != nil {
				return nil, err
			}

			gd, err := o.parseGoDoc(f.Doc.Text(), expr.file)
			if err != nil {
				return nil, err
			}

			props = append(props, jsonordered.MapItem{
				Key: name,
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
	// 文件地址
	file string

	// name 是当前表达式的字段名, 如结构体中的字段.
	name string
	// 当前表达式的唯一标识, 如 github.com/zbysir/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus
	// 此值有可能为空, 如 表达式是 model.Pet 时, 还无法获得key.
	key string

	// 如果设置为noref, 则此表达式不会使用ref代替.
	noRef bool
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
				file:    g.file,
				name:    k,
				key:     "",
			}
		}
	}

	// 查找方法
	funcs, err := g.goparse.GetStructFunc(g.goparse.GetPkgFile(g.file), g.name)
	if err != nil {
		return nil
	}

	log.Infof("funcs %+v", funcs)

	fun, exist := funcs[k]
	if !exist {
		return nil
	}

	return &GoExprWithPath{
		goparse: g.goparse,
		expr:    fun.Type,
		file:    g.file,
		name:    k,
		key:     "",
	}
}

// 把任何格式的数据都转成Schema
func (o *OpenApi) anyToSchema(i interface{}) (Schema, error) {
	switch s := i.(type) {
	case *GoExprWithPath:
		return o.goAstToSchema(s)
		//return o.goAstToSchema(s.expr, s.path)
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
