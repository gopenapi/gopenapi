package usecase

import (
	"context"
	"github.com/zbysir/gopenapi/internal/model"
)

type PetUseCase interface {
	FindPetByStatus(ctx context.Context, p *model.FindPetParams) (r model.Pet, exist bool, err error)
}
