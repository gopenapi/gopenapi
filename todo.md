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

- [ ] 简化go注释语法, 可以将schema 和 params 方法的逻辑移到js中
