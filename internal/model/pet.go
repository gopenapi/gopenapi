package model

// Pet is pet model
// $:
//   testMeta: a
type Pet struct {
	// Id is Pet ID
	Id int64 `json:"id"`

	// Category Is pet category
	Category Category `json:"category"`

	PhotoUrls []string `json:"photoUrls"`

	// Name is Pet name
	Name string `json:"name"`
	// Tag is Pet Tag
	Tags []Tag `json:"tags"`

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
// 对于组合的结构，gopenapi会将它转为allOf + ref的schema。
// 不过为了方便在js中将schema转为params, 也会同时包含properties字段.
type DelPetParams struct {
	*GetPetById
	// $required: true
	ManagePwd string `json:"manage_pwd"`
}
