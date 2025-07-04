package server

import (
	"bytes"
	"fmt"
	"strings"
)

type OperationType string

const (
	OpPut    OperationType = "PUT"
	OpGet    OperationType = "GET"
	OpDelete OperationType = "DELETE"
	OpUpdate OperationType = "UPDATE"
)

type Command struct {
	operation OperationType
	key       string
	value     []byte
	oldValue  []byte
	respChan  chan []byte
}

func ParseCommand(msg []byte) (Command, error) {
	msg = bytes.TrimSpace(msg)
	parts := bytes.Split(msg, []byte(":"))

	if len(parts) < 2 {
		return Command{}, fmt.Errorf("invalid command format: insufficient parts")
	}

	operationStr := strings.ToUpper(string(parts[0]))
	operation := OperationType(operationStr)

	key := string(parts[1])

	switch operation {
	case OpPut:
		if len(parts) < 3 {
			return Command{}, fmt.Errorf("PUT command requires a value")
		}
		return newPutCommand(key, parts[2]), nil
	case OpGet:
		return newGetCommand(key), nil
	case OpDelete:
		return newDeleteCommand(key), nil
	case OpUpdate:
		if len(parts) < 4 {
			return Command{}, fmt.Errorf("UPDATE command requires newValue and oldValue")
		}
		return newUpdateCommand(key, parts[2], parts[3]), nil
	default:
		return Command{}, fmt.Errorf("unknown operation: %s", operationStr)
	}
}

func newPutCommand(key string, value []byte) Command {
	return Command{
		operation: OpPut,
		key:       key,
		value:     value,
		respChan:  make(chan []byte),
	}
}

func newGetCommand(key string) Command {
	return Command{
		operation: OpGet,
		key:       key,
		respChan:  make(chan []byte),
	}
}

func newDeleteCommand(key string) Command {
	return Command{
		operation: OpDelete,
		key:       key,
		respChan:  make(chan []byte),
	}
}

func newUpdateCommand(key string, oldValue, newValue []byte) Command {
	return Command{
		operation: OpUpdate,
		key:       key,
		value:     newValue,
		oldValue:  oldValue,
		respChan:  make(chan []byte),
	}
}
