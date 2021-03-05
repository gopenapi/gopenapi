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

- [x] 覆盖生成的yaml

- 简化go注释语法 
  - [x] 将schema 和 params 方法的逻辑移到js中
  - [x] yaml中也可以写类似js的语法: {schema: "schema({code: 1})"}, 大多数情况够用了, 而不用必须指定是js
  - [ ] 猜测是否是js
  
- [ ] 编写IDE插件 支持在注释中语言注入yaml. 参考https://github.com/clutcher/comments_highlighter/blob/master/src/main/java/com/clutcher/comments/annotator/CommentHighlighterAnnotator.java

- [ ] Gopenapi的定位是对openapi的扩展, 应支持更多的功能(如简化redoc的x-tagGroups语法), 而通过x-$path/x-$schema等语法调用Go注释的功能只能算是一个插件.
