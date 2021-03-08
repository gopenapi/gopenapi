package cmd

const defaultConfig = "// buildin module: go that can parse definition path to schema.\nimport go from 'go';\n\nexport default {\n  filter: function (key, value) {\n    switch (key) {\n      case 'x-$path': {\n        value = go.parse(value)\n        let responses = parseResponses(value.meta.response)\n        let params = parseParams(value.meta.params)\n        let body = parseBody(value.meta.body)\n\n        let path = {\n          summary: value.summary,\n          description: value.description,\n        }\n\n        if (value.meta.tags) {\n          if (typeof value.meta.tags === 'string') {\n            path.tags = value.meta.tags.split(',').map(i => i.trim())\n          } else {\n            path.tags = value.meta.tags\n          }\n        }\n\n        if (params) {\n          path.parameters = params\n        }\n        if (body) {\n          path.requestBody = body\n        }\n        path.responses = responses\n\n        if (value.meta.security) {\n          path.security = value.meta.security.map((i) => {\n            // for 'security: [token]\n            if (typeof i === 'string') {\n              return {[i]: []}\n            } else {\n              // for 'security: [{token:write}]'\n              return i\n            }\n          })\n        }\n\n        return path\n      }\n      case 'x-$schema': {\n        value = go.parse(value)\n        return processSchema(value.schema, {omitRef: true})\n      }\n      case 'x-$tags': {\n        // for x-tagGroups syntax of redoc\n        let tagGroupsMap = {}\n        value.forEach((i) => {\n          if (i.group) {\n            if (tagGroupsMap[i.group]) {\n              tagGroupsMap[i.group].tags.push(i.name)\n            } else {\n              tagGroupsMap[i.group] = {tags: [i.name]}\n            }\n\n            delete (i.group)\n          }\n        })\n\n        let tagGroups = []\n        for (const k in tagGroupsMap) {\n          tagGroups.push({\n            name: k,\n            tags: tagGroupsMap[k].tags\n          })\n        }\n        return {\n          tags: value,\n          'x-tagGroups': tagGroups\n        }\n      }\n    }\n\n    console.warn('uncased key: ', key)\n    return value\n  }\n}\n\n// 格式化为 openApi支持的responses格式, 支持的入参格式:\n// - model.X\n// - {schema: model.X, desc: ''}\n// - schema(any)\n// - #400\n// - {200: xxx(上方三个语法), 400: xxx}\nfunction parseResponses(r) {\n  if (!r) {\n    return {\n      \"200\": {\n        $ref: '#/components/responses/200',\n      }\n    }\n  }\n  // key全部是数字\n  let keys = Object.keys(r);\n  let allIsInt = keys.length !== 0 && keys.findIndex(i => {\n    return isNaN(parseInt(i))\n  }) === -1\n\n  if (r['x-gostruct']) {\n    // case for model.X\n    return {\n      \"200\": {\n        description: 'success',\n        content: {\n          'application/json': {\n            schema: processSchema(r.schema),\n          }\n        }\n      }\n    }\n  } else if (r['x-schema']) {\n    // case for schema(model.X)\n    return {\n      \"200\": {\n        description: 'success',\n        content: {\n          'application/json': {\n            schema: processSchema(r),\n          }\n        }\n      }\n    }\n  } else if (r.schema && r.schema['x-gostruct']) {\n    // case for {schema: model.X, desc: ''}\n    return {\n      \"200\": {\n        description: r.desc || 'success',\n        content: {\n          'application/json': {\n            schema: processSchema(r.schema.schema),\n          }\n        }\n      }\n    }\n  } else if (r.schema && r.schema['x-schema']) {\n    // case for {schema: schema(model.X), desc: ''}\n    return {\n      \"200\": {\n        description: r.desc || 'success',\n        content: {\n          'application/json': {\n            schema: processSchema(r.schema),\n          }\n        }\n      }\n    }\n  } else if (typeof r === 'string') {\n    if (r[0] === '#' && r[1] !== '/') {\n      // for `#404`\n      return {\"200\": {$ref: '#/components/responses/' + r.substr(1)}}\n    } else {\n      // for `#/components/responses/404`\n      return {\"200\": {$ref: r}}\n    }\n  } else if (allIsInt) {\n    // case for {200: xxx, 400: xxx}\n    let rsp = {}\n    keys.forEach(k => {\n      let ro = parseResponses(r[k]);\n      if (ro) {\n        rsp[k] = ro[\"200\"]\n      } else {\n        console.warn(\"can't parse '\", JSON.stringify(r[k], null, 4), \"' to response\")\n      }\n    })\n\n    return rsp\n  } else {\n    console.warn(\"unexpect type of response: \", JSON.stringify(r))\n  }\n}\n\n// 格式化为openApi支持的parameters, 支持的入参格式有:\n// - []  - 数组, 则原封不动\n// - model.X  - 将schema转为params\n// - {schema: model.DelPetParams, required: ['id']}\nfunction parseParams(r) {\n  if (!r) {\n    return null\n  }\n\n  if (Array.isArray(r)) {\n    // case for []\n    return r\n  }\n\n  if (r[\"x-gostruct\"]) {\n    // case for model.X\n    if (r.schema) {\n      let properties\n      if (r.schema.type === 'object') {\n        properties = r.schema.properties\n      } else if (r.schema.allOf) {\n        // allOf语法\n        properties = r.schema['x-properties']\n      }\n\n      if (properties) {\n        let parmas = []\n        for (let k in properties) {\n          let v = properties[k]\n\n          let name = k\n          if (v.tag) {\n            if (v.tag['uri']) {\n              name = v.tag['uri']\n            } else if (v.tag['form']) {\n              name = v.tag['form']\n            } else if (v.tag['json']) {\n              name = v.tag['json']\n            }\n\n            if (name === \"-\") {\n              continue\n            }\n\n            delete (v['tag'])\n          }\n\n          let xin = 'query'\n\n          if (r.meta && r.meta['in']) {\n            xin = r.meta['in']\n          } else if (v.meta && v.meta['in']) {\n            xin = v.meta['in']\n          }\n\n          let required = null\n          if (r.meta && r.meta['required']) {\n            required = r.meta['required']\n          } else if (v.meta && v.meta['required']) {\n            required = v.meta['required']\n          }\n\n          // console.log('v 2', JSON.stringify(v))\n\n          let description = v.schema.description;\n          delete v.schema.description\n          let item = {\n            name: name,\n            description: description,\n            schema: processSchema(v.schema),\n            in: xin,\n          };\n\n          if (required !== null) {\n            item.required = required\n          }\n\n          parmas.push(item)\n        }\n        return parmas\n      }\n    }\n  } else if (r.schema && r.schema[\"x-gostruct\"]) {\n    // for {schema: model.DelPetParams, required: ['id']}\n    if (r.schema.schema) {\n      let properties\n      if (r.schema.schema.type === 'object') {\n        properties = r.schema.schema.properties\n      } else if (r.schema.schema.allOf) {\n        // allOf语法\n        properties = r.schema.schema['x-properties']\n      }\n\n      if (properties) {\n        let parmas = []\n        for (let k in properties) {\n          let v = properties[k]\n\n          let name = k\n          if (v.tag) {\n            if (v.tag) {\n              if (v.tag['uri']) {\n                name = v.tag['uri']\n              } else if (v.tag['form']) {\n                name = v.tag['form']\n              } else if (v.tag['json']) {\n                name = v.tag['json']\n              }\n              if (name === \"-\") {\n                continue\n              }\n\n              delete (v['tag'])\n            }\n\n            delete (v['tag'])\n          }\n\n          let xin = 'query'\n\n          if (r.meta && r.meta['in']) {\n            xin = r.meta['in']\n          } else if (v.meta && v.meta['in']) {\n            xin = v.meta['in']\n          }\n\n          let required = null\n          if (r.meta && r.meta['required']) {\n            required = r.meta['required']\n          } else if (v.meta && v.meta['required']) {\n            required = v.meta['required']\n          } else if (r.required) {\n            if (r.required.indexOf(name) !== -1) {\n              required = true\n            }\n          }\n\n          let description = v.schema.description;\n          delete v.schema.description\n          let item = {\n            name: name,\n            description: description,\n            schema: processSchema(v.schema),\n            in: xin,\n          };\n\n          if (required !== null) {\n            item.required = required\n          }\n\n          parmas.push(item)\n        }\n        return parmas\n      }\n    }\n  }\n\n  console.warn(\"unexpect type of params: \", JSON.stringify(r, null, 4))\n}\n\n// 格式化为openApi支持的requestBody, 支持的入参格式有:\n// - model.X\n// - schema(any)\n// - {schema: model.X, desc: \"desc\", required: ['id']}\nfunction parseBody(r) {\n  if (!r) {\n    return null\n  }\n  if (r['x-gostruct']) {\n    // for model.Pet\n    return {\n      description: 'body',\n      content: {\n        'application/json': {\n          schema: processSchema(r.schema),\n        }\n      }\n    }\n  }\n  if (r['x-schema']) {\n    // for schema(model.Pet)\n    // 不推荐的写法\n    return {\n      description: 'body',\n      content: {\n        'application/json': {\n          schema: processSchema(r),\n        }\n      }\n    }\n  } else if (r['schema'] && r['schema']['x-gostruct']) {\n    // for {schema: model.Pet, required: ['id']}\n\n    let schema\n\n    // 处理 required\n    // 语法如: {schema: model.Pet, required: ['id']}\n    if (r.required && r.required.length) {\n      // 对于指定了required值, 则不能使用ref语法\n      // note: 实际上也可以使用$ref语法, 但需要结合 allOf关键字使用, 由于swagger文档没有写这种用法, 所以还是不用$ref了.\n      schema = processSchema(r.schema.schema, {omitRef: true});\n      schema.required = r.required\n    } else {\n      schema = processSchema(r.schema.schema);\n    }\n\n    if (schema.properties && r.ext) {\n      schema.properties = Object.assign(schema.properties, r.ext)\n    }\n    return {\n      description: r.desc || 'body',\n      content: {\n        [r.bodySchema || 'application/json']: {\n          schema: schema,\n        }\n      }\n    }\n  } else if (r['schema'] && r['schema']['x-schema']) {\n    // for {schema: schema(1), desc: \"desc\"}\n    let schema = processSchema(r.schema);\n\n    // add extra properties, e.g.\n    //   {ext: {file: {type: string, format: binary}}}\n    if (schema.properties && r.ext) {\n      schema.properties = Object.assign(schema.properties, r.ext)\n    }\n\n    return {\n      description: r.desc || 'body',\n      content: {\n        [r.bodySchema || 'application/json']: {\n          schema: schema,\n        }\n      }\n    }\n  }\n}\n\n// processSchema process go-schema to openapi-schema.\n// 注意s就算是$ref, 也包含了完整的定义, 这是为了方便在js中制定更多逻辑.\n// options: {omitRef: true则忽略$ref定义, 返回全部定义, false则只返回$ref}\nfunction processSchema(s, options) {\n  if (!s) {\n    return null\n  }\n\n  // 忽略ref意味着删除$ref值, 而是返回全部值.\n  if (options && options.omitRef) {\n    if (s.$ref) {\n      delete s.$ref\n    }\n  } else {\n    if (s.$ref) {\n      return {$ref: s.$ref}\n    }\n  }\n\n  if (s.allOf) {\n    s.allOf = s.allOf.map((item) => {\n      return processSchema(item)\n    })\n    delete s['x-properties']\n  }\n\n  if (s.properties) {\n    let p = {}\n    Object.keys(s.properties).forEach(function (key) {\n      let v = s.properties[key]\n      let name = key\n\n      if (v.tag) {\n        if (v.tag.json) {\n          name = v.tag.json\n          if (name === '-') {\n            // omit this property\n            return\n          }\n        }\n        delete (v['tag'])\n      }\n\n      if (v.meta) {\n        if (v.meta.format) {\n          v.schema.format = v.meta.format\n        }\n      }\n\n      p[name] = processSchema(v.schema)\n    })\n\n    s.properties = p\n  }\n\n  if (s.items) {\n    s.items = processSchema(s.items)\n  }\n\n  if (s['x-schema']) {\n    delete s['x-schema']\n  }\n\n  if (s['x-any']) {\n    delete s['x-any']\n    // add 'example' property to fix bug of editor.swagger.io\n    if (!s.example) {\n      s.example = null\n    }\n  }\n\n  return s\n}\n"

