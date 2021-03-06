package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gopenapi/gopenapi/internal/model"
	"github.com/gopenapi/gopenapi/internal/usecase"
)

type PetHandler struct {
	u usecase.PetUseCase
}

// FindPetByStatus test for return array schema
//
// $:
//   params: model.FindPetByStatusParams
//   response:
//     200: {desc: '成功', schema: "schema([model.Pet])"}
//     401: "#401"
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

// GetPet test for pure js
//
// Returns a single pet
// $:
//    params: "[{name: 'id', required: true, in: 'path', schema: {type: 'string'}}]"
//    response:
//      200: model.Pet
//      404: {desc: 'Not Found Pet', schema: "schema({msg: 'Not Found Pet'})"}
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

// PutPet test for 'requestBody' and add custom attribute: 'required'
//
// $:
//    body: {schema: model.Pet, required: [id]}
//    response: {desc: "返回新的Pet", schema: model.Pet}
func (h *PetHandler) PutPet(ctx *gin.Context) {
	var p model.Pet
	err := ctx.ShouldBind(&p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	err = h.u.UpdatePet(context.TODO(), &p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}
	ctx.JSON(200, "ok")
}

// DelPet test for 'go-composition' syntax
//
// $:
//    params: {schema: model.DelPetParams, required: ['id']}
func (h *PetHandler) DelPet(ctx *gin.Context) {
	var p model.DelPetParams
	err := ctx.ShouldBind(&p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	err = h.u.DeletePet(context.TODO(), p.Id)
	if err != nil {
		ctx.JSON(400, err)
		return
	}
	ctx.JSON(200, "ok")
}
