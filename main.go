package main

import (
	"github.com/aokabi/ngraphinx/v2/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}

