package modelt

import "github.com/zbysir/gopenapi/internal/model"

type Pet struct {
	// Id is The Pet Id
	Id int64 `json:"id"`

	// CreateAt Is Pet's birthday
	// $format: time
	CreateAt int64 `json:"create_at"`

	// SoldAt is when the Pet is sold
	// $format: time
	SoldAt int64 `json:"sold_at"`

	// PetStatus is a test for Enum
	PetStatus PetStatus `json:"pet_status"`

	// Sex is a test for Enum
	Sex Sex `json:"sex"`

	// Category is a test for import other pkg
	Category model.Category
}

type PetStatus string

// 仅支持当前包中声明的变量/常量被识别为 Enum
// 仅支持基础类型(string, int+, float+)识别为Enum
const (
	AvailablePet PetStatus = "available"
	PendingPet   PetStatus = "pending"
	SoldPet      PetStatus = "sold"
)

type Sex byte

const (
	UnSetSex Sex = 0
	MenSex   Sex = 1
	WomenSex Sex = 2
)

var (
	AvailablePetX PetStatus = "availablex"
	PendingPetX   PetStatus = "pendingx"
	SoldPetX      PetStatus = "soldx"
)
