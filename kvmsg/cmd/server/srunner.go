package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"kvmsg/config"
	"kvmsg/internal/server"
)

func main() {
    portStr, err := config.GetServerPort()
    if err != nil {
        fmt.Printf("Failed to load port config: %v", err)
    }

    port, err := strconv.Atoi(portStr)
    if err != nil {
        fmt.Printf("Invalid port format: %v\n", err)
        os.Exit(0)
    }
	kvServer := server.New()
	err = kvServer.Start(port)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server started on port %d\n", port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	fmt.Println("\nShutting down server...")

	fmt.Printf("Active connections: %d\n", kvServer.CountActive())
	fmt.Printf("Dropped connections: %d\n", kvServer.CountDropped())

	kvServer.Close()
	fmt.Println("Server shutdown complete")
}
