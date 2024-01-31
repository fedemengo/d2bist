package main

import (
	"os"

	"github.com/fedemengo/d2bist/cmd"
	"github.com/pkg/profile"
)

func main() {
	if os.Getenv("PROFILE") == "true" {
		defer profile.Start().Stop()
	}

	cmd.Run()
}
