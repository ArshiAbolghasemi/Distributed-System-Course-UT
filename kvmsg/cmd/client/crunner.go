package main

import (
	"bufio"
	"fmt"
	"kvmsg/config"
	"kvmsg/internal/client"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	doneReading := make(chan struct{})
	doneInput := make(chan struct{})
	
	go handleRead(kvClient, stopChan, doneReading)
	go handleClientCommands(kvClient, stopChan, doneInput)

	select {
	case <-signalChan:
		fmt.Println("\nReceived shutdown signal")
	case <-stopChan:
        fmt.Println("Stopping on request...")
	}

    close(stopChan)

	<-doneReading
	<-doneInput
	fmt.Println("Closing connection...")
}

func handleRead(kvClient client.KVClient, stopChan chan struct{}, done chan struct{}) {
	for {
		select {
		case <-stopChan:
			done <- struct{}{}
			return
		default:
			resp, err := kvClient.ReadResponse()
			if err != nil {
				if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "closed") {
					fmt.Printf("\nConnection to server closed\n")
					done <- struct{}{}
					return
				}
				fmt.Printf("\nError reading from server: %v\n", err)
				continue
			}
			fmt.Printf("\nServer: %s\n> ", resp)
		}
	}
}

func handleClientCommands(kvClient client.KVClient, stopChan chan struct{}, done chan struct{}) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		select {
		case <-stopChan:
			done <- struct{}{}
			return
		default:
			command := scanner.Text()
			if strings.ToLower(command) == "exit" {
				fmt.Println("Exiting...")
				done <- struct{}{}
				return
			}
			err := kvClient.SendMessage(command)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			}
			fmt.Print("> ")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		close(stopChan)
	}
	done <- struct{}{}
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
	fmt.Println("Type 'exit' to quit the program")
}
