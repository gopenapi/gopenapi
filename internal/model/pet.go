package model

type FindPetParams struct {
	// Status values that need to be considered for filter
	Status []PetStatus `form:"status"`
}

type PetStatus string

const (
	AvailablePet PetStatus = "available"
	PendingPet   PetStatus = "pending"
	SoldPet      PetStatus = "sold"
)

type Pet struct {
	Id       int64     `json:"id"`
	Category Category  `json:"category"`
	Name     string    `json:"name"`
	Tags     []Tag     `json:"tags"`
	Status   PetStatus `json:"status"`
}
