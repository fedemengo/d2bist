package main

import (
	"github.com/fedemengo/d2bist/cmd"
)

func main() {

	//p := profile.New(
	//	profile.CPUProfile,
	//	profile.MemProfile,
	//	profile.TraceProfile,
	//)

	//defer p.Start().Stop()

	cmd.Run()
}
