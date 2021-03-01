package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gopenapi/gopenapi/internal/model"
)

type OtherHandler struct {
}

// TestRecursion
// $:
//   body: "js: model.TestRecursion"
//   response: "js: model.TestRecursion"
func (h *OtherHandler) TestRecursion(ctx *gin.Context) {
	// make sure import "model" pkg
	var _ model.TestRecursion
	return
}

// add 'tag' field
//
// $:
//   params: {schema: model.FindPetByStatusParams, required: [status]}
//   response: {200: {schema: model.Pet, desc: 'success!'}, 401: '#401'}
//   tags: [pet]
func (h *OtherHandler) Boo(ctx *gin.Context) {
	// make sure import "model" pkg
	var _ model.TestRecursion
	return
}
