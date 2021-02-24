package usecase

import (
	"context"
	"github.com/gopenapi/gopenapi/internal/model"
)

type PetUseCase interface {
	FindPetByStatus(ctx context.Context, p *model.FindPetByStatusParams) (r []model.Pet, err error)
	GetPet(ctx context.Context, p *model.GetPetById) (r model.Pet, exist bool, err error)
	UpdatePet(ctx context.Context, p *model.Pet) (err error)
	DeletePet(ctx context.Context, id int64) (err error)
}
