package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/zbysir/gopenapi/internal/pkg/goast"
	"github.com/zbysir/gopenapi/internal/pkg/gosrc"
	"github.com/zbysir/gopenapi/internal/pkg/js"
	"github.com/zbysir/gopenapi/internal/pkg/jsonordered"
	"github.com/zbysir/gopenapi/internal/pkg/log"
	"go/ast"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

type OpenApi struct {
	goparse *goast.GoParse

	// js config
	jsConfig string

	// schemas 存放需要refs的schemas
	// github.com/zbysir/gopenapi/internal/model.Category => Category
	schemas map[string]string
}

var defJsConfig = `

// {type, properties}
function pSchema(s){
  if (s.properties){
    var p ={}
	  Object.keys(s.properties).forEach(function (key) {  
		var v = s.properties[key]
		var name = key
	
		if (v.tag) {
			if (v.tag.json){
				name = v.tag.json
			}
			delete(v['tag'])
		}
		
		p[name] =  pSchema(v)
	  })
	  
	  s.properties = p
  }
  
  if (s.items){
    s.items = pSchema(s.items)
  }

  return s
}

  var config = {
      filter: function(key, value){
        if (key==='x-$path'){
          var responses = {}
          Object.keys(value.meta.resp).forEach(function (k){
            var v = value.meta.resp[k]
			var rsp 
			if (typeof v == 'string'){
				rsp = {$ref: v}
			}else{
				rsp = {
				  description: v.desc || 'success',
				  content: {
					'application/json': {
					  schema: pSchema(v.schema),
					}
				  }
				} 
			}
			
            responses[k] = rsp
          })
          return {
            parameters: value.meta.params.map(function (i){
              var x = i
              if (x.tag) {
                if (x.tag.form){
                  x.name = x.tag.form
                }
                delete(x['tag'])
              }
              if (x['meta']) {
                x.in = x['meta'].in;
				x.required = x['meta'].required
              }
              if (!x.in){
			    x.in = 'query'	
			  }

              delete(x['_from'])
              delete(x['doc'])
              delete(x['meta'])
              return x
            }),
            responses: responses
          }
        }
        if (key==='x-$schema'){
          return pSchema(value.schema)
        }

        return value
      }
    }

`

func NewOpenApi(gomodFile string, jsFile string) (*OpenApi, error) {
	goSrc, err := gosrc.NewGoSrcFromModFile(gomodFile)
	if err != nil {
		return nil, err
	}
	p := goast.NewGoParse(goSrc)

	jsConfig := defJsConfig
	if jsFile != "" {
		bs, err := ioutil.ReadFile(jsFile)
		if err != nil {
			return nil, fmt.Errorf("load jsConfig err: %w", err)
		}
		jsConfig = string(bs)
	}

	newCode, _, err := js.Transform(jsConfig, jsFile)
	if err != nil {
		return nil, fmt.Errorf("transform jsConfig to ES5 err: %w", err)
	}

	return &OpenApi{
		goparse:  p,
		jsConfig: newCode,
		schemas:  map[string]string{},
	}, nil
}

// PkgGetter 实现了 GetMember 接口, 用来给js解析器执行 member 语法.
type PkgGetter struct {
	goparse *goast.GoParse
	pkg     *goast.Pkg
}

func (p PkgGetter) GetMember(k string) interface{} {
	def, exist, err := p.goparse.GetDef(p.pkg.Dir, k)
	if err != nil {
		fmt.Printf("[err] %v", err)
		return nil
	}
	if !exist {
		return &NotFoundGoExpr{
			key: k,
			pkg: p.pkg.Dir,
		}
	}
	return &GoExprWithPath{
		goparse: p.goparse,
		expr:    def.Type,
		doc:     def.Doc,
		file:    def.File,
		name:    def.Name,
		key:     def.Key,
		noRef:   false,
	}
}

func (p PkgGetter) GetStruct(k string) (def *goast.Def, exist bool, err error) {
	return p.goparse.GetDef(p.pkg.Dir, k)
}

// 扩展openapi语法, 让其支持从go文件中读取注释信息
// params:
//  code: js表达式
//  goFilePath: 当前go文件的路径, 会根据当前文件引入的包识别js表达式使用的是哪个包.
// return:
//  可能是任何东西
func (o *OpenApi) runJsExpress(code string, goFilePath string) (interface{}, error) {
	//return nil,nil
	v, err := js.RunJs(code, func(name string) (interface{}, error) {
		// builtin function:
		// - params: for parameters of openapi
		// - schema: for schemas of openapi
		// - body: for requestBody of openapi
		switch name {
		case "params":
			return func(args ...interface{}) (interface{}, error) {
				stru := args[0]
				switch s := stru.(type) {
				case nil:
					return nil, nil
				case *GoExprWithPath:
					// case for params(model.FindPetByStatusParams) syntax:

					// 由于 运行js 表达式只支持基础类型, 所以需要转为[]interface{}
					// 又由于 go 的map[string]interface{} 是无序的, 所以不能使用copyToBaseType函数递归将所有字段都转为baseType.
					list := o.struct2ParamsList(s)
					x := make([]interface{}, len(list))
					for i, v := range list {
						x[i] = v
					}
					return x, nil
				}
				// case for non-go syntax. e.g.:
				//   params([{name: 'status', schema: 'string'}])
				return stru, nil
			}, nil
		case "schema":
			return func(args ...interface{}) (interface{}, error) {
				stru := args[0]
				return o.anyToSchema(stru)
			}, nil
		default:
			// 获取当前文件所有引入的包
			pkgs, err := o.goparse.GetFileImportedPkgs(goFilePath)
			if err != nil {
				return nil, err
			}

			// 判断name是否是pkg
			if pkg, ispkg := pkgs[name]; ispkg {
				// 如果是pkg, 则进入解析go源码流程
				return &PkgGetter{
					goparse: o.goparse,
					pkg:     pkg,
				}, nil
			}
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
func (o *OpenApi) struct2ParamsList(ep *GoExprWithPath) []ParamsItem {
	var l []ParamsItem
	switch s := ep.expr.(type) {
	case *ast.StructType:
		// 父级的 meta会传递到所有子字段
		parentGoDoc, err := o.parseGoDoc(ep.doc.Text(), ep.file)
		if err != nil {
			fmt.Printf("[err] %v", err)
			return nil
		}

		for _, f := range s.Fields.List {
			// 获取子字段 key
			name := f.Names[0].Name
			schema, err := o.goAstToSchema(&GoExprWithPath{
				goparse: o.goparse,
				expr:    f.Type,
				doc:     f.Doc,
				file:    ep.file,
				name:    name,
				key:     "",
				noRef:   false,
			})
			if err != nil {
				fmt.Printf("[err] %v\n", err)
				continue
			}

			gd, err := o.parseGoDoc(f.Doc.Text(), ep.file)
			if err != nil {
				fmt.Printf("[err] %v", err)
				continue
			}

			l = append(l, ParamsItem{
				From:        "go",
				Name:        name,
				Tag:         encodeTag(f.Tag),
				Description: gd.FullDoc,
				Meta:        mergeJsonMap(parentGoDoc.Meta, gd.Meta),
				Schema:      schema,
			})
		}
	default:
		panic(fmt.Sprintf("uncased struct2ParamsList type: %T, %+v", s, s))
	}

	return l
}

// all type of openapi: array, boolean, integer, number , object, string
func IsBaseType(t string) (is bool, openApiType string) {
	switch t {
	case "int64", "int32", "int8", "int", "uint8", "uint32", "uint64", "uint":
		return true, "integer"
	case "bool":
		return true, "boolean"
	case "float32", "float64":
		return true, "number"
	case "byte":
		return true, "integer"
	case "string":
		return true, "string"
	case "complex128":
		// 复数 暂不处理
		return true, "complex"
	}
	return
}

// 这里没有case []interface, 可能会出现问题
func yamlItemToJsonItem(i []yaml.MapItem) []jsonordered.MapItem {
	r := make([]jsonordered.MapItem, len(i))
	for i, item := range i {
		switch v := item.Value.(type) {
		case []yaml.MapItem:
			r[i] = jsonordered.MapItem{
				Key: item.Key.(string),
				Val: yamlItemToJsonItem(v),
			}
		default:
			r[i] = jsonordered.MapItem{
				Key: item.Key.(string),
				Val: v,
			}
		}
	}
	return r
}

func innerJsonToYaml(i interface{}) interface{} {
	switch i := i.(type) {
	case []jsonordered.MapItem:
		x := make([]yaml.MapItem, len(i))
		for index, item := range i {
			x[index] = yaml.MapItem{
				Key:   item.Key,
				Value: innerJsonToYaml(item.Val),
			}
		}

		return x
	case jsonordered.MapSlice:
		x := make([]yaml.MapItem, len(i))
		for index, item := range i {
			x[index] = yaml.MapItem{
				Key:   item.Key,
				Value: innerJsonToYaml(item.Val),
			}
		}
		return x
	case []interface{}:
		x := make([]interface{}, len(i))
		for index, item := range i {
			x[index] = innerJsonToYaml(item)
		}
		return x
	}

	return i
}

func jsonItemToYamlItem(i []jsonordered.MapItem) []yaml.MapItem {
	r := make([]yaml.MapItem, len(i))
	for i, item := range i {
		switch v := item.Val.(type) {
		case []jsonordered.MapItem, jsonordered.MapSlice, []interface{}:
			r[i] = yaml.MapItem{
				Key:   item.Key,
				Value: innerJsonToYaml(v),
			}
		default:
			r[i] = yaml.MapItem{
				Key:   item.Key,
				Value: v,
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
func (o *OpenApi) fullCommentMeta(i []yaml.MapItem, filename string) ([]yaml.MapItem, error) {
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
					return nil, fmt.Errorf("run js express fail: %w", err)
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
			v, err := o.fullCommentMeta(v, filename)
			if err != nil {
				return nil, err
			}
			r = append(r, yaml.MapItem{
				Key:   item.Key,
				Value: v,
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

	return r, nil
}

// 入口
// pathAndKey: e.g. github.com/zbysir/gopenapi/internal/model.Tag
func (o *OpenApi) getGoStruct(pathAndKey string, yamlKeyRouter []string) (g *GoStruct, exist bool, err error) {
	p, k := splitPkgPath(pathAndKey)

	def, exist, err := o.goparse.GetDef(p, k)
	if err != nil {
		err = fmt.Errorf("GetDef error: %w", err)
		return
	}
	if !exist {
		return
	}

	g, err = o.parseGoDoc(def.Doc.Text(), def.File)
	if err != nil {
		err = fmt.Errorf("parseGoDoc error: %w", err)
		return
	}

	// 如果是一个结构体, 则自动转为schema
	if _, ok := def.Type.(*ast.StructType); ok {
		expr := &GoExprWithPath{
			goparse: o.goparse,
			expr:    def.Type,
			doc:     def.Doc,
			file:    def.File,
			name:    def.Name,
			key:     def.Key,
			noRef:   false,
		}

		if isSchemasComponentsKey(yamlKeyRouter) {
			expr.noRef = true
		}
		g.Schema, err = o.goAstToSchema(expr)
		if err != nil {
			err = fmt.Errorf("toSchema error: %w", err)
			return
		}
	}

	return
}

// splitPkgPath 分割路径src 为path和包名
//  pathAndKey: ../internal/pkg/goast.GoMeta
//  output:
//    path: ../internal/pkg/goast
//    key: GoMeta
func splitPkgPath(src string) (pa, member string) {
	p1, filename := path.Split(src)
	ss := strings.Split(filename, ".")
	if len(ss) != 0 {
		p1 += ss[0]
		member = strings.Join(ss[1:], ".")
	}

	pa = p1
	return
}

type XData struct {
	Summary     string
	Description string

	Meta map[string]interface{}
}

// 从go结构体能读出的数据, 用于parameters
type ParamsItem struct {
	// From 表示此item来至那, 如 go
	From        string               `json:"_from"`
	Name        string               `json:"name"`
	Tag         map[string]string    `json:"tag"`
	Description string               `json:"description"`
	Meta        jsonordered.MapSlice `json:"meta,omitempty"`
	Schema      Schema               `json:"schema"`
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
		Value: t.Description,
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
func (o *OpenApi) CompleteYaml(inYaml string) (dest string, err error) {
	// 读取openapi
	var kv []yaml.MapItem

	err = yaml.Unmarshal([]byte(inYaml), &kv)
	if err != nil {
		return "", err
	}

	err = o.walkSchemas(kv)
	if err != nil {
		return "", err
	}

	newKv, err := o.completeYaml(kv, []string{})
	if err != nil {
		return
	}

	out, err := yaml.Marshal(newKv)
	if err != nil {
		return
	}

	dest = string(out)
	return
}

func walkYamlItem(kv []yaml.MapItem, wantKeys []string, walkedKeys []string, cb func(key []string, i yaml.MapItem)) {
	for _, item := range kv {
		key := ""

		switch k := item.Key.(type) {
		case string:
			key = k
		case int:
			key = strconv.Itoa(k)
		default:
			panic(fmt.Sprintf("uncase Type of itemKey: %T", item.Key))
		}

		if key == wantKeys[0] || wantKeys[0] == "*" {
			if len(wantKeys) == 1 {
				cb(walkedKeys, item)
				break
			}

			switch val := item.Value.(type) {
			case []yaml.MapItem:
				walkYamlItem(val, wantKeys[1:], append(walkedKeys, key), cb)
			}
		}
	}
}

func (o *OpenApi) walkSchemas(kv []yaml.MapItem) (err error) {
	walkYamlItem(kv, []string{"components", "schemas", "*", "x-$schema"}, nil, func(key []string, i yaml.MapItem) {
		schemaKey := strings.Join(key, "/")
		pat, ok := i.Value.(string)
		if !ok {
			return
		}

		pat, inProject := o.goparse.FormatPath(pat)
		if !inProject {
			return
		}
		o.schemas[pat] = schemaKey
	})
	return
}

// isSchemasComponentsKey 返回key是否是定义schema的key
func isSchemasComponentsKey(key []string) bool {
	if len(key) < 2 {
		return false
	}
	return key[0] == "components" && key[1] == "schemas"
}

func (o *OpenApi) runConfigJs(key string, in []byte, keyRouter []string) (jsBs []byte, err error) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	console.Enable(vm)

	_, err = vm.RunScript("builtin", "var exports= {};")
	if err != nil {
		return
	}
	_, err = vm.RunScript("jsConfig", o.jsConfig)
	if err != nil {
		err = fmt.Errorf("RunScript err: %w", err)
		return
	}

	krBs, _ := json.Marshal(keyRouter)
	code := fmt.Sprintf(`var r = exports.default.filter("%s", %s, %s); JSON.stringify(r)`, key, in, krBs)
	//log.Infof("%s", code)
	v, err := vm.RunScript("js", code)
	if err != nil {
		err = fmt.Errorf("RunScript err: %w", err)
		return
	}

	//log.Infof("%T %v", v.Export(),v.Export())
	exp := v.Export()
	if exp == nil {
		return []byte(`{}`), nil
	}
	s := v.Export().(string)

	return []byte(s), err
}

// keyRoute: key的路径
func (o *OpenApi) completeYaml(in []yaml.MapItem, keyRouter []string) (out []yaml.MapItem, err error) {
	for _, item := range in {
		key := ""

		switch k := item.Key.(type) {
		case string:
			key = k
		case int:
			key = strconv.Itoa(k)
		default:
			panic(fmt.Sprintf("uncase Type of itemKey: %T", item.Key))
		}

		// x-$xxx 语法, 将调用go注释
		if strings.HasPrefix(key, "x-$") {
			if v, ok := item.Value.(string); ok {
				g, exist, err := o.getGoStruct(v, keyRouter)
				if err != nil {
					err = fmt.Errorf("full yaml '%s' fail\n  %w", strings.Join(keyRouter, "."), err)
					return nil, err
				}
				if !exist {
					log.Warningf("error at %s : can't resolve path: %s", strings.Join(keyRouter, "."), v)
					out = append(out, yaml.MapItem{
						Key:   key,
						Value: fmt.Sprintf("gopenapi-err, can't resolve path: %s", v),
					})
					continue
				}

				inbs, err := json.Marshal(g)
				if err != nil {
					return nil, err
				}

				outBs, err := o.runConfigJs(key, inbs, keyRouter)
				if err != nil {
					return nil, err
				}

				// 让json有序
				var outI jsonordered.MapSlice

				err = json.Unmarshal(outBs, &outI)
				if err != nil {
					return nil, err
				}

				bsxxx, _ := json.Marshal(outI)
				if !bytes.Equal(bsxxx, outBs) {
					log.Errorf("Unexpected result on Marshal, want: %s, got: %s", outBs, bsxxx)
				}
				x := jsonItemToYamlItem(outI)
				out = append(out, x...)
				continue
			}
		}

		switch v := item.Value.(type) {
		// TODO 注意是否有[]interface类型
		//case []interface{}:
		case []yaml.MapItem:
			completeYaml, err := o.completeYaml(v, append(keyRouter, key))
			if err != nil {
				return nil, err
			}
			out = append(out, yaml.MapItem{
				Key:   item.Key,
				Value: completeYaml,
			})
			continue
		}

		out = append(out, item)
	}
	return
}
