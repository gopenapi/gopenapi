package main

import "github.com/zbysir/gopenapi/internal/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
