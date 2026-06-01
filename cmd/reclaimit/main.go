package main

import (
	"os"

	"github.com/svg153/reclaimit"
)

func main() {
	os.Exit(reclaimit.Run(os.Args[1:], os.Stdout, os.Stderr))
}
