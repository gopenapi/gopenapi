export default {
  // 格式化 responseItem, 支持的参数格式:
  // - model.X
  // - {schema: model.X, desc: ''}
  // - #400
  // - {200: xxx(上方三个语法), 400: xxx}

  parseResponseItem: function (r) {
    if (!r) {
      return {
        "200": {
          $ref: '#/components/responses/200',
        }
      }
    }
    // key全部是数字
    let allIsInt = Object.keys(r).length !== 0 && Object.keys(r).findIndex(i => {
      return parseInt(i) === 0
    }) === -1

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
        rsp[k] = this.parseResponseItem(r[k])["200"]
      })

      return rsp
    } else {
      console.warn("unexpect type of response: ", JSON.stringify(r))
    }
  },
  // 格式化params, 支持的格式有:
  // - []: 数组, 则原封不动
  // - GoStruct: 将schema转为params
  // - {schema: model.DelPetParams, required: ['id']}
  parseParams: function (r) {
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
              if (v.tag['form']) {
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
              if (v.tag['form']) {
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
  },

  // 格式化requestBody
  // - GoStruct
  // - schema
  // - {schema: model.X, desc: "desc"}
  parseBody: function (r) {
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

  },
  filter: function (key, value) {

    switch (key) {
      case 'x-$path': {
        let responses = this.parseResponseItem(value.meta.response)

        // console.log('params 1', JSON.stringify(value.meta.params))
        let params = this.parseParams(value.meta.params)


        let body = this.parseBody(value.meta.body)
        // console.log('value.meta.body', JSON.stringify(value.meta.body))

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

// processSchema process go-schema to openapi-schema.
// 注意s就算是$ref, 则包含了完整的定义, 这是为了方便在js中制定更多逻辑.
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
