export default {
  filter: function (key, value) {
    switch (key) {
      case 'x-$path': {
        let responses = parseResponses(value.meta.response)
        let params = parseParams(value.meta.params)
        let body = parseBody(value.meta.body)

        let path = {
          summary: value.summary,
          description: value.description,
        }

        if (params) {
          path.parameters = params
        }
        if (body) {
          path.requestBody = body
        }

        path.responses = responses

        return path
      }
      case 'x-$schema': {
        return processSchema(value.schema)
      }
    }
    return value
  }
}

// 格式化为 openApi支持的responses格式, 支持的入参格式:
// - model.X
// - {schema: model.X, desc: ''}
// - schema(any)
// - #400
// - {200: xxx(上方三个语法), 400: xxx}
function parseResponses(r) {
  if (!r) {
    return {
      "200": {
        $ref: '#/components/responses/200',
      }
    }
  }
  // key全部是数字
  let keyLen = Object.keys(r).length;
  let allIsInt = keyLen !== 0 && Object.keys(r).filter(i => {
    return parseInt(i) > 0
  }).length === keyLen

  if (r['x-gostruct']) {
    // case for model.X
    return {
      "200": {
        description: 'success',
        content: {
          'application/json': {
            schema: processSchema(r.schema),
          }
        }
      }
    }
  } else if (r['x-schema']) {
    // case for schema(model.X)
    return {
      "200": {
        description: 'success',
        content: {
          'application/json': {
            schema: processSchema(r),
          }
        }
      }
    }
  } else if (r.schema && r.schema['x-gostruct']) {
    // case for {schema: model.X, desc: ''}
    return {
      "200": {
        description: r.desc || 'success',
        content: {
          'application/json': {
            schema: processSchema(r.schema.schema),
          }
        }
      }
    }
  } else if (r.schema && r.schema['x-schema']) {
    // case for {schema: schema(model.X), desc: ''}
    return {
      "200": {
        description: r.desc || 'success',
        content: {
          'application/json': {
            schema: processSchema(r.schema),
          }
        }
      }
    }
  } else if (typeof r === 'string') {
    if (r[0] === '#' && r[1] !== '/') {
      // for `#404`
      return {"200": {$ref: '#/components/responses/' + r.substr(1)}}
    } else {
      // for `#/components/responses/404`
      return {"200": {$ref: r}}
    }
  } else if (allIsInt) {
    // case for {200: xxx, 400: xxx}
    let rsp = {}
    Object.keys(r).forEach(k => {
      let ro = parseResponses(r[k]);
      if (ro) {
        rsp[k] = ro["200"]
      } else {
        console.warn("can't parse '", JSON.stringify(r[k], null, 4), "' to response")
      }
    })

    return rsp
  } else {
    console.warn("unexpect type of response: ", JSON.stringify(r))
  }
}

// 格式化为openApi支持的parameters, 支持的入参格式有:
// - []  - 数组, 则原封不动
// - model.X  - 将schema转为params
// - {schema: model.DelPetParams, required: ['id']}
function parseParams(r) {
  if (!r) {
    return null
  }

  if (Array.isArray(r)) {
    // case for []
    return r
  }

  if (r["x-gostruct"]) {
    // case for model.X
    if (r.schema) {
      let properties
      if (r.schema.type === 'object') {
        properties = r.schema.properties
      } else if (r.schema.allOf) {
        // allOf语法
        properties = r.schema['x-properties']
      }

      if (properties) {
        let parmas = []
        for (let k in properties) {
          let v = properties[k]

          let name = k
          if (v.tag) {
            if (v.tag['uri']) {
              name = v.tag['uri']
            } else if (v.tag['form']) {
              name = v.tag['form']
            } else if (v.tag['json']) {
              name = v.tag['json']
            }

            delete (v['tag'])
          }

          let xin = 'query'

          if (r.meta && r.meta['in']) {
            xin = r.meta['in']
          } else if (v.meta && v.meta['in']) {
            xin = v.meta['in']
          }

          let required = null
          if (r.meta && r.meta['required']) {
            required = r.meta['required']
          } else if (v.meta && v.meta['required']) {
            required = v.meta['required']
          }

          // console.log('v 2', JSON.stringify(v))

          let description = v.schema.description;
          delete v.schema.description
          let item = {
            name: name,
            description: description,
            schema: processSchema(v.schema),
            in: xin,
          };

          if (required !== null) {
            item.required = required
          }

          parmas.push(item)
        }
        return parmas
      }
    }
  } else if (r.schema && r.schema["x-gostruct"]) {
    // for {schema: model.DelPetParams, required: ['id']}
    if (r.schema.schema) {
      let properties
      if (r.schema.schema.type === 'object') {
        properties = r.schema.schema.properties
      } else if (r.schema.schema.allOf) {
        // allOf语法
        properties = r.schema.schema['x-properties']
      }

      if (properties) {
        let parmas = []
        for (let k in properties) {
          let v = properties[k]

          let name = k
          if (v.tag) {
            if (v.tag) {
              if (v.tag['uri']) {
                name = v.tag['uri']
              } else if (v.tag['form']) {
                name = v.tag['form']
              } else if (v.tag['json']) {
                name = v.tag['json']
              }

              delete (v['tag'])
            }

            delete (v['tag'])
          }

          let xin = 'query'

          if (r.meta && r.meta['in']) {
            xin = r.meta['in']
          } else if (v.meta && v.meta['in']) {
            xin = v.meta['in']
          }

          let required = null
          if (r.meta && r.meta['required']) {
            required = r.meta['required']
          } else if (v.meta && v.meta['required']) {
            required = v.meta['required']
          } else if (r.required) {
            if (r.required.indexOf(name) !== -1) {
              required = true
            }
          }

          // console.log('v 2', JSON.stringify(v))

          let description = v.schema.description;
          delete v.schema.description
          let item = {
            name: name,
            description: description,
            schema: processSchema(v.schema),
            in: xin,
          };

          if (required !== null) {
            item.required = required
          }

          parmas.push(item)
        }
        return parmas
      }
    }
  }

  console.warn("unexpect type of params: ", JSON.stringify(r, null, 4))
}

// 格式化为openApi支持的requestBody, 支持的入参格式有:
// - model.X
// - schema(any)
// - {schema: model.X, desc: "desc", required: ['id']}
function parseBody(r) {
  if (!r) {
    return null
  }
  if (r['x-gostruct']) {
    // for model.Pet
    return {
      description: 'body',
      content: {
        'application/json': {
          schema: processSchema(r.schema),
        }
      }
    }
  }
  if (r['x-schema']) {
    // for schema(model.Pet)
    // 不推荐的写法
    return {
      description: 'body',
      content: {
        'application/json': {
          schema: processSchema(r),
        }
      }
    }
  } else if (r['schema'] && r['schema']['x-gostruct']) {
    // for {schema: model.Pet, required: ['id']}

    let schema

    // 处理 required
    // 语法如: {schema: model.Pet, required: ['id']}
    if (r.required && r.required.length) {
      // 对于指定了required值, 则不能使用ref语法
      // note: 实际上也可以使用$ref语法, 但需要结合 allOf关键字使用, 由于swagger文档没有写这种用法, 所以还是不用$ref了.
      schema = processSchema(r.schema.schema, {omitRef: true});
      schema.required = r.required
    } else {
      schema = processSchema(r.schema.schema);
    }
    return {
      description: r.desc || 'body',
      content: {
        'application/json': {
          schema: schema,
        }
      }
    }
  } else if (r['schema'] && r['schema']) {
    // for {schema: schema(1), desc: "desc"}
    return {
      description: r.desc || 'body',
      content: {
        'application/json': {
          schema: processSchema(r.schema),
        }
      }
    }
  }
}

// processSchema process go-schema to openapi-schema.
// 注意s就算是$ref, 也包含了完整的定义, 这是为了方便在js中制定更多逻辑.
// options: {omitRef: true则忽略$ref定义, 返回全部定义, false则只返回$ref}
function processSchema(s, options) {
  if (!s) {
    return null
  }

  // 忽略ref意味着删除$ref值, 而是返回全部值.
  if (options && options.omitRef) {
    if (s.$ref) {
      delete s.$ref
    }
  } else {
    if (s.$ref) {
      return {$ref: s.$ref}
    }
  }
  if (s.modify) {
    console.log('modify: ', JSON.stringify(s))
  }

  if (s.allOf) {
    s.allOf = s.allOf.map((item) => {
      return processSchema(item)
    })
    delete s['x-properties']
  }

  if (s.properties) {
    let p = {}
    Object.keys(s.properties).forEach(function (key) {
      let v = s.properties[key]
      var name = key

      if (v.tag) {
        if (v.tag.json) {
          name = v.tag.json
          if (name === '-') {
            // omit this property
            return
          }
        }
        delete (v['tag'])
      }
      // console.log('v.schema 2', JSON.stringify(v.schema))

      p[name] = processSchema(v.schema)

      // 处理 required
      // if (s.modify) {
      //   s.modify.forEach(({k, v}) => {
      //     if (k === "required") {
      //       if (v.indexOf(key) !== -1) {
      //         p[name].required = true
      //       }
      //     }
      //   })
      // }
    })

    s.properties = p
  }

  if (s.items) {
    s.items = processSchema(s.items)
  }

  if (s['x-schema']) {
    delete s['x-schema']
  }

  if (s['x-any']) {
    delete s['x-any']
    // add 'example' property to fix bug of editor.swagger.io
    if (!s.example) {
      s.example = null
    }
  }

  return s
}
