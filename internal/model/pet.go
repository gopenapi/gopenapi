package model

type Pet struct {
	// Id is Pet ID
	//Id       int64     `json:"id"`

	// Category Is pet category
	Category Category `json:"category"`

	// Id is Pet name
	//Name     string    `json:"name"`
	// Tag is Pet Tag
	// $a: 1
	//Tags     []Tag     `json:"tags"`

	// PetStatus
	Status PetStatus `json:"status"`
}

type Pets []Pet

type PetStatus string

const (
	AvailablePet PetStatus = "available"
	PendingPet   PetStatus = "pending"
	SoldPet      PetStatus = "sold"
)

// $in: path
type FindPetByStatusParams struct {
	// Status values that need to be considered for filter
	// $:
	//   required: true
	//   in: query
	Status []PetStatus `form:"status"`
}

type GetPetById struct {
	// Id of pet to return
	// $:
	//   required: true
	//   in: path
	Id int64 `uri:"id"`
}
