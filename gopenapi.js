// {type, properties}
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

var config = {
  filter: function (key, value) {
    if (key === 'x-$path') {
      var responses = {}
      Object.keys(value.meta.resp).forEach(function (k) {
        var v = value.meta.resp[k]
        responses[k] = {
          description: v.desc || 'success',
          content: {
            'application/json': {
              schema: pSchema(v.schema),
            }
          }
        }
      })
      return {
        parameters: value.meta.params.map(function (i) {
          var x = i
          delete (x['_from'])
          if (x.tag) {
            if (x.tag.form) {
              x.name = x.tag.form
            }
            delete (x['tag'])
          }
          delete (x['meta'])
          delete (x['doc'])
          if (!x.in) {
            x.in = 'query'
          }

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
