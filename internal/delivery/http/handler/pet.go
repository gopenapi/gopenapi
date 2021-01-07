package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/zbysir/gopenapi/internal/model"
	"github.com/zbysir/gopenapi/internal/usecase"
)

type PetHandler struct {
	u usecase.PetUseCase
}

// FindPet Is Api that do Finds Pets by status
//
// Multiple status values can be provided with comma separated strings
// goparams: query {...model.FindPetParams}
// goresponse: json [model.Pet]
// or goresponse: json [#/components/schemas/Pet]
func (h *PetHandler) FindPet(ctx *gin.Context) {
	var p model.FindPetParams
	err := ctx.ShouldBind(&p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	r, exist, err := h.u.FindPetByStatus(context.TODO(), &p)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	if !exist {
		ctx.JSON(404, "not found")
	}

	ctx.JSON(200, r)
}
