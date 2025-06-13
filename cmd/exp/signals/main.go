package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Using Notify/Stop (have to make a channel)
	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, os.Interrupt)
	// // Wait for the signal
	// <- ch
	// // Stop receiving signals
	// signal.Stop(ch)

	// Using NotifyContext
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	// Wait for context to be cancelled via timeout or for signal
	<- ctx.Done()
	// Stop receiving signals
	stop()

	// Simulate a slow shutdown
	fmt.Println("Shutting down...")
	time.Sleep(3 * time.Second)
	fmt.Println("Shutdown complete")
}
