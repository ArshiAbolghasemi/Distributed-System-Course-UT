package client

import (
	"bufio"
	"fmt"
	"kvmsg/config"
	"net"
	"strings"
)

type KVC struct {
	conn   net.Conn
	reader *bufio.Reader
	done   chan struct{}
}

func New() *KVC {
	return &KVC{
		done: make(chan struct{}),
	}
}

func (c *KVC) Connect(address string) error {
    config, err := config.LoadConfig("../../config/config.yml")
    if err != nil {
        return err
    }
	conn, err := net.Dial(config.Server.Protocol, address)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

func (c *KVC) SendMessage(message string) error {	
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	
	_, err := c.conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	
	return nil
}

func (c *KVC) ReadResponse() (string, error) {
	response, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	
	return strings.TrimSpace(response), nil
}

func (c *KVC) Close() {
	if c.conn != nil {
		c.conn.Close()
		close(c.done)
	}
}
