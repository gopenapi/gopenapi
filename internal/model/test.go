package model

type TestRecursion struct {
	Id       int64            `json:"id"`
	Children []*TestRecursion `json:"children"`
}
