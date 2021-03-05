package main

import "github.com/gopenapi/gopenapi/internal/cmd"

//go:generate go run ./internal/cmd/gen/main.go ./gopenapi.conf.js ./internal/cmd/gen.go cmd defaultConfig

func main() {
	_ = cmd.Execute()
}
