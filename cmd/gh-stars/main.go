package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/psauxwwf/gh-stars/internal/stars"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := stars.Run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "gh-stars:", err)
		os.Exit(1)
	}
}
