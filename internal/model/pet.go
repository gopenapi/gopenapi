package model

// Pet is pet model
// $:
//   testMeta: a
type Pet struct {
	//Id is Pet ID
	Id int64 `json:"id"`

	// Category Is pet category
	Category Category `json:"category"`

	// Name is Pet name
	Name string `json:"name"`
	// Tag is Pet Tag
	Tags []Tag `json:"tags"`

	// PetStatus
	Status PetStatus `json:"status"`

	// test interface
	X interface{} `json:"x"`
}

type Pets []Pet

type PetStatus string

const (
	AvailablePet PetStatus = "available"
	PendingPet   PetStatus = "pending"
	SoldPet      PetStatus = "sold"
)

// $in: query
type FindPetByStatusParams struct {
	// Status values that need to be considered for filter
	// $required: true
	Status []PetStatus `form:"status"`
}

type GetPetById struct {
	// Id of pet to return
	// $:
	//   required: true
	//   in: path
	Id int64 `uri:"id"`
}

// DelPetParams test for allOf syntax
// 对于组合的结构，gopenapi会尝试使用allOf + ref语法。
// TODO: 不过为了方便在js中将schema转为params, 也会同时生产完整的objectSchema结构.
type DelPetParams struct {
	*Pet
	ManagePwd string `json:"manage_pwd"`
}
