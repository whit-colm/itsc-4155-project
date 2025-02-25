package main

import (
	"os"

	"github.com/whit-colm/itsc-4155-project/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
