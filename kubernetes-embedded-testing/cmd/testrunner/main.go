package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	configFile     string
	projectRoot    string
	debug          bool
	workspacePath  string
	image          string
	testCommand    string
	keepNamespace  bool
	backoffLimit   int32
	activeDeadline int64
	testExitCode   int
	logPrefix      bool
	logTimestamp   bool
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandling(ctx, cancel)

	rootCmd := createRootCommand(ctx)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if testExitCode != 0 {
			os.Exit(testExitCode)
		}
		os.Exit(1)
	}

	os.Exit(0)
}

func setupSignalHandling(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()
}
