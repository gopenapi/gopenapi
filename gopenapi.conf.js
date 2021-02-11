export default {
  filter: function (key, value) {
    switch (key) {
      case 'x-$path': {
        let responses = {}
        if (value.meta.response) {
          if (value.meta.response['x-schema']) {
            // for schema(xxx) syntax
            responses = {
              "200": {
                description: 'success',
                content: {
                  'application/json': {
                    schema: processSchema(value.meta.response),
                  }
                }
              }
            }
          } else if (value.meta.response.schema) {
            // for `{desc: "xxx", schema: schema(xxx)}` syntax
            responses = {
              "200": {
                description: value.meta.response.desc || 'success',
                content: {
                  'application/json': {
                    schema: processSchema(value.meta.response.schema),
                  }
                }
              }
            }
          } else {
            // for {200: xxx} syntax
            Object.keys(value.meta.response).forEach(function (k) {
              let v = value.meta.response[k]
              let rsp
              if (typeof v == 'string') {
                if (v[0] === '#' && v[1] !== '/') {
                  // for `#404`
                  rsp = {$ref: '#/components/responses/' + v.substr(1)}
                } else {
                  // for `#/components/responses/404`
                  rsp = {$ref: v}
                }
              } else {
                let schema
                if (v['x-schema']) {
                  schema = v
                } else {
                  schema = v.schema
                }
                rsp = {
                  description: v.desc || 'success',
                  content: {
                    'application/json': {
                      schema: processSchema(schema),
                    }
                  }
                }
              }

              responses[k] = rsp
            })
          }
        } else {
          // add default response
          responses = {
            "200": {
              $ref: '#/components/responses/200',
            }
          }
        }

        let params
        // console.log('params', JSON.stringify(value.meta.params))
        if (value.meta.params) {
          params = []

          value.meta.params.forEach(function (i) {
            let x = i
            if (x.tag) {
              if (x.tag.form) {
                x.name = x.tag.form
              } else if (x.tag.json) {
                x.name = x.tag.json
              }

              if (x.name === '-') {
                // omit this property
                return
              }
            }
            if (x['meta']) {
              x.in = x['meta'].in;
              x.required = x['meta'].required
            }
            if (!x.in) {
              // default `in` in openapi-parameters
              x.in = 'query'
            }
            if (x.schema) {
              x.schema = processSchema(x.schema)
            }

            delete (x['tag'])
            delete (x['_from'])
            delete (x['meta'])

            params.push(x)
          })
        }

        let body
        if (value.meta.body) {
          if (value.meta.body['x-schema']) {
            // for schema(xxx) syntax
            body = {
              description: 'body',
              content: {
                'application/json': {
                  schema: processSchema(value.meta.body),
                }
              }
            }
          } else {
            let v = value.meta.body
            let schema
            schema = v.schema

            body = {
              description: v.desc || 'body',
              content: {
                'application/json': {
                  schema: processSchema(schema),
                }
              }
            }
          }
        }

        return {
          summary: value.summary,
          description: value.description,
          parameters: params,
          requestBody: body,
          responses: responses,
        }
      }
      case 'x-$schema': {
        return processSchema(value.schema)
      }
    }
    return value
  }
}

// processSchema process go-schema to openapi-schema.
function processSchema(s) {
  if (!s) {
    return null
  }

  if (s.allOf) {
    s.allOf = s.allOf.map((item) => {
      return processSchema(item)
    })
  }

  if (s.properties) {
    var p = {}
    Object.keys(s.properties).forEach(function (key) {
      var v = s.properties[key]
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

      if (v['x-any']){
        delete v['x-any']
        // add 'example' property to fix bug of editor.swagger.io
        if (!v.example){
          v.example = null
        }
      }

      p[name] = processSchema(v)
    })

    s.properties = p
  }

  if (s.items) {
    s.items = processSchema(s.items)
  }

  if (s['x-schema']) {
    delete s['x-schema']
  }

  return s
}
