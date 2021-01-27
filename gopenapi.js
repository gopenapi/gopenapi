function pSchema(s) {
  if (s.properties) {
    var p = {}
    Object.keys(s.properties).forEach(function (key) {
      var v = s.properties[key]
      var name = key

      if (v.tag) {
        if (v.tag.json) {
          name = v.tag.json
        }
        delete (v['tag'])
      }

      p[name] = pSchema(v)
    })

    s.properties = p
  }

  if (s.items) {
    s.items = pSchema(s.items)
  }

  return s
}

export default {
  filter: function (key, value) {
    if (key === 'x-$path') {
      let responses = {}
      Object.keys(value.meta.resp).forEach(function (k) {
        let v = value.meta.resp[k]
        let rsp
        if (typeof v == 'string') {
          rsp = {$ref: v}
        } else {
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
        parameters: value.meta.params.map(function (i) {
          let x = i
          if (x.tag) {
            if (x.tag.form) {
              x.name = x.tag.form
            }
            delete (x['tag'])
          }
          if (x['meta']) {
            x.in = x['meta'].in;
            x.required = x['meta'].required
          }
          if (!x.in) {
            x.in = 'query'
          }

          delete (x['_from'])
          delete (x['doc'])
          delete (x['meta'])
          return x
        }),
        responses: responses
      }
    }
    if (key === 'x-$schema') {
      return pSchema(value.schema)
    }

    return value
  }
}
