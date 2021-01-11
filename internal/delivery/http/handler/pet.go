package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/zbysir/gopenapi/internal/model"
	"github.com/zbysir/gopenapi/internal/usecase"
)

// a 111
var (
	// a 222
	a = 1
	// b bbb
	b = 1
)

// PetHandler doc
type PetHandler struct {
	u usecase.PetUseCase
}

// Test type doc
type (
	PetHandler2 struct {
		u usecase.PetUseCase
	}
)

// FindPetByStatus Is Api that do Finds Pets by status
// .abc
//
// Multiple status values can be provided with comma separated strings
//
// $path
//   params: "js: [...model.FindPetByStatusParams, {name: status, required: true}]"
//   resp: 'js: {200: {desc: "成功", content: [model.Pet]}, 401: {desc: "没权限", content: {msg: "没权限"}}}'
//
func (h *PetHandler) FindPetByStatus(ctx *gin.Context) {
	var p model.FindPetByStatusParams
	err := ctx.ShouldBind(&p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	r, err := h.u.FindPetByStatus(context.TODO(), &p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	ctx.JSON(200, r)
}

// $path
//    params: {...model.GetPetById}
//    resp: model.Pet
//
func (h *PetHandler) GetPet(ctx *gin.Context) {
	var p model.GetPetById
	err := ctx.ShouldBindUri(&p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	r, exist, err := h.u.GetPet(context.TODO(), &p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	if !exist {
		ctx.JSON(404, "not found")
	}

	ctx.JSON(200, r)
}
