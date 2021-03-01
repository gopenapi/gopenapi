package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/gopenapi/gopenapi/internal/pkg/goast"
	"github.com/gopenapi/gopenapi/internal/pkg/jsonordered"
	"github.com/gopenapi/gopenapi/internal/pkg/log"
	"go/ast"
	"sort"
)

type Schema interface {
	_schema()
	setRef(ref string) Schema
}

var _ Schema = &ErrSchema{}
var _ Schema = &ObjectSchema{}
var _ Schema = &ArraySchema{}
var _ Schema = &AllOfSchema{}
var _ Schema = &AnySchema{}
var _ Schema = &IdentSchema{}

type ObjectSchema struct {
	Ref         string               `json:"$ref,omitempty"`
	Type        string               `json:"type"`
	Description string               `json:"description,omitempty"`
	Properties  jsonordered.MapSlice `json:"properties"`
	Example     interface{}          `json:"example,omitempty"`

	Modify   []Modify `json:"modify,omitempty"`
	IsSchema bool     `json:"x-schema,omitempty"`
}

func (o *ObjectSchema) setRef(ref string) Schema {
	o.Ref = ref
	return o
}

// 实现装饰器语法
// schema(model.x).require('id','name', any)
func (o *ObjectSchema) GetMember(k string) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		o.Modify = append(o.Modify, Modify{
			Key:  k,
			Args: args,
		})
		return o, nil
	}, nil
}

func (o *ObjectSchema) _schema() {}

type ArraySchema struct {
	Ref         string `json:"$ref,omitempty"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Items       Schema `json:"items"`
	IsSchema    bool   `json:"x-schema"`
}

func (a *ArraySchema) setRef(ref string) Schema {
	a.Ref = ref
	return a
}

func (a *ArraySchema) _schema() {}

//type RefSchema struct {
//	Ref      string `json:"$ref"`
//	IsSchema bool   `json:"x-schema"`
//}

type Modify struct {
	Key  string
	Args []interface{}
}

//func (r *RefSchema) _schema() {}

// 基础类型, string / int
type IdentSchema struct {
	Ref string `json:"$ref,omitempty"`

	Type        string        `json:"type"`
	Description string        `json:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
	IsSchema    bool          `json:"x-schema,omitempty"`

	Example interface{} `json:"example,omitempty"`
}

func (s *IdentSchema) _schema() {}

func (s *IdentSchema) setRef(ref string) Schema {
	s.Ref = ref
	return s
}

type ErrSchema struct {
	IsSchema bool `json:"x-schema,omitempty"`
	// 用于强提示，此字段在editor中会报错。
	Error string `json:"error,omitempty"`
	// 用于弱提示，此字段在editor中不会报错。
	XError string `json:"x-error,omitempty"`
}

func (n ErrSchema) setRef(ref string) Schema {
	panic("implement me")
}

func (n ErrSchema) _schema() {
}

type AnySchema struct {
	IsSchema    bool          `json:"x-schema,omitempty"`
	IsAny       bool          `json:"x-any,omitempty"`
	Description string        `json:"description,omitempty"`
	OneOf       []interface{} `json:"oneOf"`
}

func (n AnySchema) setRef(ref string) Schema {
	panic("implement me")
}

func (n AnySchema) _schema() {
}

type AllOfSchema struct {
	AllOf []Schema `json:"allOf"`

	// AllOf 也有Properties字段, 这是为了在js中使用此字段转为params数组
	Properties jsonordered.MapSlice `json:"x-properties"`
	IsSchema   bool                 `json:"x-schema"`
}

func (n AllOfSchema) _schema() {
}

func (n AllOfSchema) setRef(ref string) Schema {
	panic("implement me")
}

// ObjectProp 对象的成员
type ObjectProp struct {
	Schema Schema               `json:"schema"`
	Meta   jsonordered.MapSlice `json:"meta,omitempty"`
	Tag    map[string]string    `json:"tag,omitempty"`
}

//type Expr struct {
//	// 表达式, 如 model.Pet
//	expr ast.Expr
//	// 当前表达式所在的文件, 当expr是 model.Pet 时, 会根据当前文件的import找到model包下的Pet结构体.
//	exprInFile string
//	key string
//}

// Key 返回某表达式的唯一标识.
// 用于判断是否重复, 如 ref的生成, 递归判断.
// 返回key格式如 pkgname.Ident, e.g. github.com/gopenapi/gopenapi/internal/delivery/http/handler.PetHandler
// 例:
//   - *model.Pet, 返回 xxx/model.Pet
//   - model.Pet , 返回 xxx/model.Pet
//   - Pet , 返回 {当前包}.Pet
//   - 其他情况下则认为没有唯一标识.
func (expr *GoExprWithPath) Key() (string, error) {
	return expr.key, nil
}

// goAstToSchema 将goAst转为Schema
//
//   expr参数是goAst
//   exprInFile 是这个expr在哪一个文件中(必须是相对路径, 如github.com/gopenapi/gopenapi/internal/model/pet.go), 这是为了识别到这个文件引入了哪些包.
func (o *OpenApi) goAstToSchema(expr *GoExprWithPath, noRef bool) (Schema, error) {
	ga := GoAstToSchema{
		goparse:      o.goparse,
		schemas:      o.schemas,
		parsedSchema: map[string]int{},
		openapi:      o,
	}

	return ga.goAstToSchema(expr, noRef)
}

func (o *GoAstToSchema) goAstToSchema(goExpr *GoExprWithPath, noRef bool) (Schema, error) {
	if !noRef {
		// 使用ref逻辑

		// 判断此表达式是否是在schemas中定义过了，如果定义过了则使用ref语法。
		exprKey, err := goExpr.Key()
		if err != nil {
			err = fmt.Errorf("call Expr.Key err: %w", err)
			return nil, err
		}
		if ref, ok := o.schemas[exprKey]; ok {
			return ref.schema.setRef("#/" + ref.yamlKey), nil
			//return &RefSchema{
			//	Ref:      "#/" + ref.yamlKey,
			//	IsSchema: true,
			//}, nil
		}
	}

	k, err := goExpr.Key()
	if err != nil {
		return nil, err
	}

	if k != "" {
		// 可以递归两次，超出则报错
		if count, ok := o.parsedSchema[k]; ok && count >= 2 {
			msg := fmt.Sprintf("recursive references on '%s'", k)
			return &ErrSchema{IsSchema: true, XError: msg}, nil
		}

		o.parsedSchema[k] ++
	}

	switch expr := goExpr.expr.(type) {
	case *ast.ArrayType:
		schema, err := o.goAstToSchema(&GoExprWithPath{
			goparse: o.goparse,
			openapi: o.openapi,
			expr:    expr.Elt,
			doc:     goExpr.doc,
			file:    goExpr.file,
			name:    "",
			key:     "",
		}, false)
		if err != nil {
			return nil, err
		}
		gd, err := o.openapi.parseGoDoc(goExpr.doc.Text(), goExpr.file)
		if err != nil {
			return nil, err
		}
		schema = &ArraySchema{
			Type:        "array",
			Items:       schema,
			IsSchema:    true,
			Description: gd.FullDoc,
		}
		return schema, nil
	case *ast.StarExpr:
		return o.goAstToSchema(&GoExprWithPath{
			openapi: o.openapi,
			goparse: goExpr.goparse,
			expr:    expr.X,
			doc:     goExpr.doc,
			file:    goExpr.file,
			name:    goExpr.name,
			key:     goExpr.key,
		}, false)
	case *ast.Ident:
		// 标识
		// 如果是基础类型, 则返回, 否则还需要继续递归.
		if is, t := IsBaseType(expr.Name); is {
			gd, err := o.openapi.parseGoDoc(goExpr.doc.Text(), goExpr.file)
			if err != nil {
				return nil, err
			}

			return &IdentSchema{
				Type:        t,
				Default:     nil,
				Enum:        nil,
				IsSchema:    true,
				Description: gd.FullDoc,
				Example:     nil,
			}, nil
		}
		def, exist, err := o.goparse.GetDef(o.goparse.GetPkgOfFile(goExpr.file), expr.Name)
		// 获取当前包下的结构体
		if err != nil {
			return nil, err
		}
		if !exist {
			msg := fmt.Sprintf("can't found Type: %s", expr.Name)
			// TODO print error
			log.Warning(msg)
			return &ErrSchema{
				Error: msg,
			}, nil
		}

		doc := goExpr.doc
		if doc == nil {
			doc = def.Doc
		}
		schema, err := o.goAstToSchema(&GoExprWithPath{
			openapi: o.openapi,
			goparse: o.goparse,
			expr:    def.Type,
			doc:     doc,
			file:    def.File,
			name:    "",
			key:     def.Key,
		}, false)
		if err != nil {
			return nil, err
		}

		// 如果是**子级是**基础类型, 则需要获取枚举值
		// e.g. "type Status string"代码中 需要获取Status的枚举值.
		if idt, ok := schema.(*IdentSchema); ok {
			// 查找Enum
			enum, err := o.goparse.GetEnum(o.goparse.GetPkgOfFile(goExpr.file), expr.Name)
			if err != nil {
				return nil, err
			}

			_, defValue := enum.FirstValue()

			idt.Enum = enum.Values
			idt.Default = defValue
		}

		return schema, err
	case *ast.SelectorExpr:
		// for model.T syntax
		pkgName := expr.X.(*ast.Ident).Name

		pkgs, err := o.goparse.GetFileImportedPkgs(goExpr.file)
		if err != nil {
			return nil, err
		}

		if pkg, isPkg := pkgs[pkgName]; isPkg {
			str, exist, err := PkgGetter{
				goparse: o.goparse,
				pkg:     pkg,
			}.GetStruct(expr.Sel.Name)
			if err != nil || !exist {
				return &ErrSchema{}, err
			}

			schema, err := o.goAstToSchema(&GoExprWithPath{
				openapi: o.openapi,
				goparse: o.goparse,
				doc:     str.Doc,
				expr:    str.Type,
				file:    str.File,
				name:    str.Name,
				key:     str.Key,
			}, false)
			if err != nil {
				return nil, err
			}
			return schema, err
		}

		return &ErrSchema{}, nil
	case *ast.StructType:
		var props jsonordered.MapSlice

		allOf := AllOfSchema{
			AllOf:    nil,
			IsSchema: true,
		}
		for _, f := range expr.Fields.List {
			fieldSchema, err := o.goAstToSchema(&GoExprWithPath{
				openapi: o.openapi,
				goparse: o.goparse,
				expr:    f.Type,
				file:    goExpr.file,
				name:    "",
				//name:    name,
				key: "",
				doc: f.Doc,
			}, false)
			if err != nil {
				return nil, err
			}

			var name string
			// nested
			// 组合（嵌套）格式处理
			if len(f.Names) != 0 {
				name = f.Names[0].Name
			} else if f.Tag != nil {
				name = getExprName(f.Type)
			} else {
				// 对于golang的组合语法, 都使用allOf语法实现
				//
				// 当是嵌套, 并且没有任何tag时, 才展开子级
				// 如果字段schema是refSchema，则使用allOf语法
				// 如果字段是ObjectSchema，则展开
				// 如果不是上面则情况，则当成普通字段处理
				allOf.AllOf = append(allOf.AllOf, fieldSchema)
				switch t := fieldSchema.(type) {
				case *ObjectSchema:
					allOf.Properties = append(allOf.Properties, t.Properties...)
				}
				continue
			}

			gd, err := o.openapi.parseGoDoc(f.Doc.Text(), goExpr.file)
			if err != nil {
				return nil, err
			}

			props = append(props, jsonordered.MapItem{
				Key: name,
				Val: ObjectProp{
					Schema: fieldSchema,
					Meta:   gd.Meta,
					Tag:    encodeTag(f.Tag),
				},
			})
		}
		gd, err := o.openapi.parseGoDoc(goExpr.doc.Text(), goExpr.file)
		if err != nil {
			return nil, err
		}
		var schema Schema = &ObjectSchema{
			Type:        "object",
			Properties:  props,
			IsSchema:    true,
			Description: gd.FullDoc,
			Example:     nil,
			Modify:      nil,
		}

		if len(allOf.AllOf) != 0 {
			allOf.AllOf = append(allOf.AllOf, schema)
			allOf.Properties = append(allOf.Properties, props...)
			return allOf, nil
		}

		return schema, nil
	case *ast.InterfaceType:
		gd, err := o.openapi.parseGoDoc(goExpr.doc.Text(), goExpr.file)
		if err != nil {
			return nil, err
		}

		return &AnySchema{
			IsSchema:    true,
			IsAny:       true,
			Description: gd.FullDoc,
			OneOf: []interface{}{
				map[string]interface{}{"type": "array"},
				map[string]interface{}{"type": "boolean"},
				map[string]interface{}{"type": "integer"},
				map[string]interface{}{"type": "number"},
				map[string]interface{}{"type": "object"},
				map[string]interface{}{"type": "string"},
			},
		}, nil
	default:
		panic(fmt.Sprintf("uncased goAstToSchema type: %T, %+v", expr, expr))
	}
}

// getExprName返回表达式在嵌套语法中的字段名
// e.g.
// - model.Category 返回 Category
func getExprName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getExprName(t.X)
	case *ast.SelectorExpr:
		return getExprName(t.Sel)
	default:
		panic(fmt.Sprintf("uncased type '%T' for getExprName", e))
	}
}

// NotFoundGoExpr 用于表示没有找到Go表达式
//  如 schema(apkg.xxx), 如果apkg中没有找到xxx, 则会返回 NotFoundGoExpr
//  不返回nil的原因是需要有更多的信息用于友好提示
type NotFoundGoExpr struct {
	key string
	pkg string
}

type GoExprWithPath struct {
	goparse *goast.GoParse
	openapi *OpenApi
	expr    ast.Expr
	// 表达式的文档
	doc *ast.CommentGroup
	// 表达式所在的文件地址
	file string

	// name 是当表达式是一个结构体时 结构体的名字.
	// 如 type X struct{} 中的X
	// 用于如下语法, 通过这个名字获取该结构体上的方法.
	//   func (x X) FuncA(){}
	name string
	// 当前表达式的唯一标识, 如 github.com/gopenapi/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus
	// 此值有可能为空, 如 表达式是具体的某个结构体声明时无法获得key.
	key string
}

// 在解析成json时(在js脚本中使用), 需要解析成js脚本能使用的格式, 即 GoStruct
func (g *GoExprWithPath) MarshalJSON() ([]byte, error) {
	str, err := g.openapi.parseGoDoc(g.doc.Text(), g.file)
	if err != nil {
		err = fmt.Errorf("parseGoDoc error: %w", err)
		return nil, err
	}

	sch, err := g.openapi.anyToSchema(g)
	if err != nil {
		return nil, fmt.Errorf("to schema %w", err)
	}

	str.Schema = sch

	return json.Marshal(str)
}

// 如果类型是 结构体, 则还需要查询到子方法或者子成员
func (g *GoExprWithPath) GetMember(k string) (interface{}, error) {
	// 如果类型是 结构体, 则还需要查询到子方法或者子成员
	str, ok := g.expr.(*ast.StructType)
	if !ok {
		return nil, nil
	}

	// 查找结构体中子成员
	for _, field := range str.Fields.List {
		if k == field.Names[0].Name {
			return &GoExprWithPath{
				openapi: g.openapi,
				goparse: g.goparse,
				expr:    field.Type,
				doc:     field.Doc,
				file:    g.file,
				name:    "",
				key:     "",
			}, nil
		}
	}

	// 查找方法
	funcs, err := g.goparse.GetFuncOfStruct(g.goparse.GetPkgOfFile(g.file), g.name)
	if err != nil {
		return nil, nil
	}

	fun, exist := funcs[k]
	if !exist {
		return nil, nil
	}

	return &GoExprWithPath{
		goparse: g.goparse,
		openapi: g.openapi,
		expr:    fun.Type,
		doc:     fun.Doc,
		file:    g.file,
		name:    "",
		key:     "",
	}, nil
}

type schemaSave struct {
	yamlKey string
	schema  Schema
}

type GoAstToSchema struct {
	goparse *goast.GoParse

	// 已经定义了的schema
	// key(def key in go) => yaml key(e.g. components/schema/Pet)
	schemas map[string]schemaSave

	// 已经解析过的schema, 用于判断无限递归
	parsedSchema map[string]int

	openapi *OpenApi
}

// 把任何格式的数据都转成Schema
func (o *OpenApi) anyToSchema(i interface{}) (Schema, error) {
	switch s := i.(type) {
	case *GoExprWithPath:
		return o.goAstToSchema(s, false)
	case []interface{}:
		if len(s) == 0 {
			return &ArraySchema{
				Type: "array",
				Items: &AnySchema{
					IsSchema:    true,
					IsAny:       true,
					Description: "",
					OneOf: []interface{}{
						map[string]interface{}{"type": "array"},
						map[string]interface{}{"type": "boolean"},
						map[string]interface{}{"type": "integer"},
						map[string]interface{}{"type": "number"},
						map[string]interface{}{"type": "object"},
						map[string]interface{}{"type": "string"},
					},
				},
				IsSchema: true,
			}, nil
		}
		item, err := o.anyToSchema(s[0])
		if err != nil {
			return nil, err
		}
		return &ArraySchema{
			Type:     "array",
			Items:    item,
			IsSchema: true,
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
					Schema: p,
					Meta:   nil,
					Tag:    nil,
				},
			})
		}

		return &ObjectSchema{
			Type:       "object",
			Properties: props,
			IsSchema:   true,
		}, nil
	case string:
		return &IdentSchema{
			Type:        "string",
			Default:     s,
			Enum:        nil,
			IsSchema:    true,
			Description: "",
			Example:     s,
		}, nil
	case int64, int, int8, int32, uint, uint64, uint32, uint8:
		return &IdentSchema{
			Type:     "integer",
			Default:  s,
			IsSchema: true,
			Example:  s,
		}, nil
	case float64, float32:
		return &IdentSchema{
			Type:     "number",
			Default:  s,
			IsSchema: true,
			Example:  s,
		}, nil
	case bool:
		return &IdentSchema{
			Type:     "boolean",
			Default:  s,
			IsSchema: true,
			Example:  s,
		}, nil
	case *NotFoundGoExpr:
		return &ErrSchema{
			IsSchema: true,
			Error:    fmt.Sprintf("can't found definition '%s' in pkg '%s'", s.key, s.pkg),
		}, nil
	case error:
		return &ErrSchema{
			IsSchema: true,
			Error:    s.Error(),
		}, nil
	default:
		panic(fmt.Sprintf("uncased type2Schema type: %T, %+v", s, s))
	}
}
