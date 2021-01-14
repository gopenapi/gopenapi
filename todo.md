- [ ] schema支持枚举

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

- [ ] 关联ref
