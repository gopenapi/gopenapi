- [x] schema支持枚举

  处理如下代码的 PetStatus 为基础类型, 并且生成枚举/default
  ```
    type PetStatus string
    
    const (
        AvailablePet PetStatus = "available"
        PendingPet   PetStatus = "pending"
        SoldPet      PetStatus = "sold"
    )
    
    // $in: path
    type FindPetByStatusParams struct {
        // Status values that need to be considered for filter
        // $required: true
        Status []PetStatus `form:"status"`
    }

  ```

- [x] 关联ref
  除了 components/schemas下声明的schemas, 其他引用的schemas都会使用ref关联.

- [ ] 覆盖生成的yaml

- [ ] requestBody 中 schema 中 required 的处理, 和 params 中 required 的处理.
  - 语法暂定: schema(model.X, {required: 'id'}, {xxx: xxx}), 其中后面的入参可以使用 x-ext-args 字段在js中访问.
      params(model.X, {append: {name: 'ext', in: 'path'}}, {required: 'id'}), 现在params方法将不再返回array, 而是object, 目的是为了能够访问到 x-ext-args 字段

- [ ] 简化go注释语法, 可以将schema 和 params 方法的逻辑移到js中