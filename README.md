# Gopenapi

Gopenapi helps you to get rid of the cumbersome definition when writing openapi spec, and keep it as flexible as the
native yaml syntax

## Goals

- Highly flexible
    - Base on openapi.yaml file, you can choose which part of the content is generated by Gopenapi according to your
      preferences
    - Use JavaScript to write additional logic
- Understanding low-cost
    - Syntax as simple as yaml
- Less intrusion
    - Just need a few comments in the go source code

## Why not

- [go-swagger](https://github.com/go-swagger/go-swagger) has the following disadvantages:
    - Syntax is complicated, I now need to learn two kinds of syntax: openapi and goswagger
    - Non-scalability, only the syntax supported by go-swgger can be used
    - It maks the source code is not clean enough, as time advances, the source code will become more and more verbose

## Shortcoming

- Currently, this project is still being tested, Performance is not stable enough, code is not elegant enough.
- Since it is based on Golang Ast, it currently only supports the Golang.

## Get start

### Precondition

- Make sure your project is written in Golang
- Make sure 'go module' is enabled
- Make sure the 'go.mod' file is in the project root directory

### Step 0: Install Gopenapi

```shell
go get -u github.com/gopenapi/gopenapi
```

### Step 1: Write openapi.src.yaml file

> Gopenapi only support yaml format files so far

```yaml
openapi: 3.0.1
info:
  title: Swagger Petstore
  description: |
    This is a sample server Petstore server.  You can find out more about     Swagger
    at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).      For
    this sample, you can use the api key "special-key" to test the authorization     filters.

    **All api maybe return 401 code if it need auth.**
  termsOfService: http://swagger.io/terms/
  contact:
    email: apiteam@swagger.io
  license:
    name: Apache 2.0
    url: http://www.apache.org/licenses/LICENSE-2.0.html
  version: 1.0.0
externalDocs:
  description: Find out more about Swagger
  url: http://swagger.io
servers:
  - url: https://petstore.swagger.io/v2
  - url: http://petstore.swagger.io/v2

tags:
  - name: pet
    description: Test all features

paths:
  /pet/findByStatus:
    get:
      x-$path: github.com/gopenapi/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus
      tags:
        - pet
      security:
        - petstore_auth:
            - write:pets
            - read:pets
  /pet/{id}:
    get:
      x-$path: github.com/gopenapi/gopenapi/internal/delivery/http/handler.PetHandler.GetPet
      tags:
        - pet
      operationId: getPetById
      security:
        - api_key: [ ]

    delete:
      x-$path: ./internal/delivery/http/handler.PetHandler.DelPet
      tags:
        - pet
      security:
        - api_key: [ ]

components:
  schemas:
    Category:
      x-$schema: github.com/gopenapi/gopenapi/internal/model.Category
    Tag:
      x-$schema: ./internal/model.Tag
    Pet:
      x-$schema: github.com/gopenapi/gopenapi/internal/model.Pet
```

As you can see, Gopenapi provide some extended syntax:

- x-$schema
- x-$path

#### x-$path

Its value is the definition path of function or struct.

e.g.

- github.com/gopenapi/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus
- ./internal/delivery/http/handler.PetHandler.FindPetByStatus

#### x-$schema

Its value is the definition path of struct.

e.g.

- github.com/gopenapi/gopenapi/internal/model.Pet
- ./internal/model.Pet

### Step 2: Write 'meta-comments' in your go source code

```go
// FindPetByStatus Finds Pets by status
//
// $:
//   params: model.FindPetByStatusParams
//   response: schema([model.Pet])
func (h *PetHandler) FindPetByStatus(ctx *gin.Context) {
...
}
```

**Done! Just added 3 lines of comments**

'meta-comments' syntax is yaml, but it needs to start with the '$' symbol.

The following syntax is correct:

```
// $:
//   params: model.FindPetByStatusParams
//   response: schema([model.Pet])
```

or

```
// $:
//   params: model.FindPetByStatusParams
//   response: 
//     200: schema([model.Pet])
```

or

```
// $params: model.FindPetByStatusParams
// $response: schema([model.Pet])
```

> Remember it just is 'yaml'

### Step 3: Run Gopenapi to fill your yaml file

```bash
gopenapi -i example/openapi.src.yaml -o example/openapi.gen.yaml
```

You can type 'gopenapi -h' to get more helps.

```bash
gopenapi -h
```

> Tip: You can review the generated file on http://editor.swagger.io/

## Advanced

### What is `gopenapi.conf.js`?

Gopenapi will generate a gopenapi.conf.js file in the root directory of the project after running if this file does not
exist.

This script is designed to allow you to customize the convenient syntax according to your preferences.

The following syntax are supported by current script:

```
// add 'required' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: model.Pet
```

```
// add 'description' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {schema: model.Pet, desc: 'success!'}
```

```
// add more responses
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'} 
```

```
// add 'tag' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'} 
//   tags: [pet]
```

You can write any data in 'meta-comments' and then process it into openapi data in `gopenapi.conf.js`.

Because of it, Gopenapi becomes flexible.

### How to write `gopenapi.conf.js`?

Comments in go:
```
// add 'tag' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'} 
//   tags: [pet]
```

It will be processed into the following data, and you can use it in `gopenapi.conf.js`:

```
{
    "doc": "add 'tag' field",
    "summary": "add 'tag' field",
    "description": "",
    "meta": {
        "params": {
            "schema": {
                "doc": "",
                "summary": "",
                "description": "",
                "meta": {
                    "in": "query"
                },
                "schema": {
                    "type": "object",
                    "properties": {
                        "Status": {
                            "schema": {
                                "type": "array",
                                "description": "Status values that need to be considered for filter",
                                "items": {
                                    "type": "string",
                                    "description": "Status values that need to be considered for filter",
                                    "default": "available",
                                    "enum": [
                                        "available",
                                        "pending",
                                        "sold"
                                    ],
                                    "x-schema": true
                                },
                                "x-schema": true
                            },
                            "meta": {
                                "required": true
                            },
                            "tag": {
                                "form": "status"
                            }
                        }
                    },
                    "x-schema": true
                },
                "x-gostruct": true
            },
            "required": [
                "status"
            ]
        },
        "response": {
            "200": {
                "schema": {
                    "doc": "Pet is pet model",
                    "summary": "Pet is pet model",
                    "description": "",
                    "meta": {
                        "testMeta": "a"
                    },
                    "schema": {
                        "type": "object",
                        "description": "Pet is pet model",
                        "properties": {
                            "Id": {
                                "schema": {
                                    "type": "integer",
                                    "description": "Id is Pet ID",
                                    "x-schema": true
                                },
                                "tag": {
                                    "json": "id"
                                }
                            },
                            "Category": {
                                "schema": {
                                    "type": "object",
                                    "description": "Category Is pet category",
                                    "properties": {msg: 'emit'},
                                    "x-schema": true
                                },
                                "tag": {
                                    "json": "category"
                                }
                            },
                            "Name": {
                                "schema": {
                                    "type": "string",
                                    "description": "Name is Pet name",
                                    "x-schema": true
                                },
                                "tag": {
                                    "json": "name"
                                }
                            },
                            "Status": {
                                "schema": {
                                    "type": "string",
                                    "description": "PetStatus",
                                    "default": "available",
                                    "enum": [
                                        "available",
                                        "pending",
                                        "sold"
                                    ],
                                    "x-schema": true
                                },
                                "tag": {
                                    "json": "status"
                                }
                            }
                        },
                        "x-schema": true
                    },
                    "x-gostruct": true
                },
                "desc": "success!"
            },
            "401": "#401"
        },
        "tags": [
            "pet"
        ]
    },
    "x-gostruct": true
}
```

The schema part is complicated. Fortunately, we don’t care about it, but focus on other fields. 
For example, if we want to add a `operationId` field, we can write code like this:

in go comments:
```
// add 'tag' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'} 
//   tags: [pet]
//   operationId: FindPetByStatus
```


in `gopenapi.conf.js`:
```diff
export default {
  filter: function (key, value) {
        ...
+++        if (value.meta.operationId) {
+++          path.operationId = value.meta.operationId
+++        }
        ...
```

## FQA

#### How to distinguish whether the string in 'meta-comments' is JavaScript or pure string?

```
// $:
//   params: model.FindPetByStatusParams
```

In the above example, 'model.FindPetByStatusParams' will be parsed into JavaScript execution instead of pure string,
because Gopenapi can guessing whether the string is JavaScript or not.

Strings that meet the following rules will be executed as JavaScript:

- Wrap with '[]'
- Wrap with '{}'
- Wrap with 'schema()'
- string like "model.X" and can find the definition in go source code.

## Next goal
- Optimize code and performance
- More documentation if you need it
