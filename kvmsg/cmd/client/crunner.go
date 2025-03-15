package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"kvmsg/config"
	"kvmsg/internal/client"
)

const (
	reset  = "\033[0m"
	blue   = "\033[34m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
)

func main() {
	serverAddr, err := config.GetServerAddress()
	if err != nil {
		fmt.Printf("Failed to Load Config: %v", err)
		os.Exit(1)
	}
	kvClient := client.New()
	err = kvClient.Connect(serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer kvClient.Close()

	greeting(serverAddr)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	stopChan := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-stopChan:
				return
			default:
				command := scanner.Text()

				if strings.ToLower(command) == "exit" {
					fmt.Println("Exiting...")
					close(stopChan)
					return
				}

				handleCommand(kvClient, command)
			}
		}
	}()

	select {
	case <-signalChan:
		fmt.Println("\nReceived shutdown signal")
	case <-stopChan:
	}

	fmt.Println("Closing connection...")
}

func greeting(serverAddr string) {
	fmt.Printf("%sConnected to KV server at %s%s\n\n", blue, serverAddr, reset)

	fmt.Println(green + "Available operations:" + reset)
	fmt.Println(yellow + "  GET    " + reset + "- Retrieve a value")
	fmt.Println(yellow + "  PUT    " + reset + "- Store a new key-value pair")
	fmt.Println(yellow + "  UPDATE " + reset + "- Modify an existing value")
	fmt.Println(yellow + "  DELETE " + reset + "- Remove a key")

	fmt.Println(green + "Command formats:" + reset)
	fmt.Println(red + "  Get:key" + reset)
	fmt.Println(red + "  Put:key:value" + reset)
	fmt.Println(red + "  Update:key:oldValue:newValue" + reset)
	fmt.Println(red + "  Delete:key" + reset)
}

func handleCommand(kvc client.KVClient, commandStr string) {
	err := kvc.SendMessage(commandStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	resp, err := kvc.ReadResponse()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp)
}
