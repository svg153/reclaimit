package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/svg153/reclaimit"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(reclaimit.RunContext(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
