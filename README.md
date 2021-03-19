<p align="center">
  <img alt="Gopenapi logo" src="./logo.png" height="150" />
  <h3 align="center">Gopenapi</h3>
  <p align="center">Gopenapi use javascript to extend and simplify openapi sepc.</p>
</p>

---

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
- Since it is based on Golang Ast, it currently **only supports Golang**, and need to enable 'go module'.

## Who needs Gopenapi?

If you meet the following situation, maybe you need it

- Not using openapi
- Write openapi.yaml by yourself

Don't recommend you who are using go-swagger to use it, unless you want to write `openapi.yaml` again.

## Get start

### Precondition

- Make sure your project is written in Golang and 'go module' is enabled

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

For more information about 'extended syntax', please refer to [here](#Extended-Syntax)

### Step 2: Write 'meta-comments' in your go source code

```diff
// FindPetByStatus Finds Pets by status
//
+ // $:
+ //   params: model.FindPetByStatusParams
+ //   response: schema([model.Pet])
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

Run the following command in the project root directory.
```bash
gopenapi -i example/openapi.src.yaml -o example/openapi.gen.yaml
```

Tip: Type 'gopenapi -h' for more helps.

```bash
gopenapi -h

Usage:
  gopenapi [flags]

Flags:
  -c, --config string   specify the configuration file to be used (default "gopenapi.conf.js")
  -h, --help            help for gopenapi
  -i, --input string    specify the source file in yaml format
  -o, --output string   specify the output file path
  -v, --version         version for gopenapi

```

> Tip: You can review the generated file on http://editor.swagger.io/

## Extended Syntax

#### x-$path

The x-$path instruction generates data that conforms to `openapi-path` from Go comment.

Its value is the definition path of function or struct.

e.g.

- github.com/gopenapi/gopenapi/internal/delivery/http/handler.PetHandler.FindPetByStatus
- ./internal/delivery/http/handler.PetHandler.FindPetByStatus

#### x-$schema

The x-$schema instruction generates data that conforms to `openapi-schema` from Go struct.

Its value is the definition path of struct.

e.g.

- github.com/gopenapi/gopenapi/internal/model.Pet
- ./internal/model.Pet

#### x-$tags

Same as `tags`, except that the group field is added to generate 'x-tagGroups'
for [redoc](https://github.com/Redocly/redoc).

input

```yaml

x-$tags:
  - name: Regular
    description: Test all features
    group: GroupA
  - name: Edge
    description: Test things that are prone to bugs
    group: GroupB
```

output

```yaml
tags:
  - name: Regular
    description: Test all features
  - name: Edge
    description: Test things that are prone to bugs
x-tagGroups:
  - name: GroupA
    tags:
      - Regular
  - name: GroupB
    tags:
      - Edge
```

## Advanced

So far you already know how to use Gopenapi, if you want more customization, please read on.

### What is `gopenapi.conf.js`?

Gopenapi will generate `gopenapi.conf.js` file in the root directory of the project after running if this file does not
exist.

You can make more custom syntax in this script (In addition to cross-language syntax, such as the functions provided by
x-$path, it is built in Gopenapi. nevertheless, you can still customize the format of 'meta-comment').

For example, the x-$path syntax currently supports multiple formats of 'meta-comment':

in yaml

```yaml
x-$path: ./internal/delivery/http/handler.PetHandler.FindPetByStatus
```

in go code

```
// add 'required' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: model.Pet

or

// add 'description' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {schema: model.Pet, desc: 'success!'}

or

// add more responses
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'} 

or

// add 'tag' field
// 
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'} 
//   tags: [pet]
```

### How to write `gopenapi.conf.js`?

The code in `gopenapi.conf.js` is very long, I do not recommend you to modify it, you can create a new issue if you have
same requirements.

If you really want to modify it, don’t worry about breaking it, just delete it and run gopenapi again will regenerate
it.

For example, if we want to add a `operationId` field, we can write code like this:

In go comments:

```diff
// $:
//   params: model.FindPetByStatusParams
//   response: schema([model.Pet]) 
//   tags: [pet]
+ //   operationId: FindPetByStatus
```

In `gopenapi.conf.js`:

```diff
export default {
  filter: function (key, value) {
    ...
    let path = {
      summary: value.summary,
      description: value.description,
    }
     // use 'console.log(JSON.stringify(value))' to print data-struct of value
+    if (value.meta.operationId) {
+      path.operationId = value.meta.operationId
+    }
    ...
```

Now the generated openapi will add the `operationId` field
```diff
paths:
  /pet/findByStatus:
    get:
+      operationId: FindPetByStatus
      summary: FindPetByStatus test for return array schema
      description: ""
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
- Organize the code in `gopenapi.conf.js`
