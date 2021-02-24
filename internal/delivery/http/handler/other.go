package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zbysir/gopenapi/internal/model"
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
