package main

import (
	"github.com/aokabi/ngraphinx/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}

