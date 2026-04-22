package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/halkyon/dp/testapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server, err := testapi.NewServer()
	if err != nil {
		return err
	}

	fmt.Println(server.Addr())

	return server.Run(ctx)
}
