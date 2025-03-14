package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"

	"kvmsg/config"
	"kvmsg/internal/kvstore"
)

type Client struct {
	conn        net.Conn
	id          int
	messageChan chan []byte
	done        chan struct{}
}

type KVServer struct {
	listener           net.Listener
	port               int
    kvstore            kvstore.KVStore
	started            bool
	closed             bool
	clients            map[int]*Client
	nextClientID       int
	activeCount        int
	droppedCount       int
	opsChan            chan Command
	registerClientChan chan *Client
	removeClientChan   chan int
	countChan          chan chan int
	droppedChan        chan chan int
	shutdown           chan struct{}
}

func NewKVServer() *KVServer {
	return &KVServer{
		clients:            make(map[int]*Client),
		nextClientID:       1,
        kvstore:            kvstore.New(),
		opsChan:            make(chan Command),
		registerClientChan: make(chan *Client),
		removeClientChan:   make(chan int),
		countChan:          make(chan chan int, 500),
		droppedChan:        make(chan chan int, 500),
		shutdown:           make(chan struct{}),
	}
}

func (s *KVServer) Start(port int) error {
	if s.started {
		return fmt.Errorf("server already started")
	}

    config, err := config.LoadConfig("../../config/config.yml")
    if err != nil {
        return err
    }
	listener, err := net.Listen(config.Server.Protocol, config.GetServerAddress())
	if err != nil {
		return err
	}

	s.listener = listener
	s.port = port
	s.started = true
    s.closed = false

	// Start KV store handler
	go s.handleKVStore()

	// Start client manager
	go s.manageClients()

	// Accept connections
	go s.acceptConnections()

	return nil
}

func (s *KVServer) manageClients() {
	for {
		select {
		case client := <-s.registerClientChan:
			s.clients[client.id] = client
			s.activeCount++

		case clientID := <-s.removeClientChan:
			if _, exists := s.clients[clientID]; exists {
				delete(s.clients, clientID)
				s.activeCount--
				s.droppedCount++
			}

		case respChan := <-s.countChan:
			respChan <- s.activeCount

		case respChan := <-s.droppedChan:
			respChan <- s.droppedCount

		case <-s.shutdown:
			for _, client := range s.clients {
				close(client.messageChan)
				client.conn.Close()
				close(client.done)
			}
			return
		}
	}
}

func (s *KVServer) handleKVStore() {
	for {
		select {
		case cmd := <-s.opsChan:
			switch cmd.operation {
			case OpPut:
				s.kvstore.Put(cmd.key, cmd.value)
				cmd.respChan <- []byte("OK")

			case OpGet:
				values := s.kvstore.Get(cmd.key)
				if len(values) == 0 {
					cmd.respChan <- []byte("NOT_FOUND")
				} else {
					result := bytes.Join(values, []byte("\n"))
					cmd.respChan <- result
				}

			case OpDelete:
				s.kvstore.Delete(cmd.key)
				cmd.respChan <- []byte("OK")

			case OpUpdate:
				s.kvstore.Update(cmd.key, cmd.oldValue, cmd.value)
				cmd.respChan <- []byte("OK")
			}

		case <-s.shutdown:
			return
		}
	}
}

func (s *KVServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return
			default:
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
		}

        config, err := config.LoadConfig("../../config/config.yml")
        if err != nil {
            select {
			case <-s.shutdown:
				return
			default:
				fmt.Printf("%v\n", err)
				continue
			}
        }

		client := &Client{
			conn:        conn,
			id:          s.nextClientID,
			messageChan: make(chan []byte, config.Client.ChannelBufferLimit),
			done:        make(chan struct{}),
		}
		s.nextClientID++

		s.registerClientChan <- client

		// Handle client in goroutines
		go s.handleClientRead(client)
		go s.handleClientWrite(client)
	}
}

// handleClientRead reads messages from the client
func (s *KVServer) handleClientRead(client *Client) {
	reader := bufio.NewReader(client.conn)

	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from client %d: %v\n", client.id, err)
			}
			s.removeClientChan <- client.id
			close(client.done)
			return
		}

        cmd, err := ParseCommand(message)
        s.opsChan <- cmd

		resp := <-cmd.respChan
		resp = append(resp, '\n')

		select {
		case client.messageChan <- resp:
			// Message sent to client's buffer
		default:
			// Buffer full, drop message
			fmt.Printf("Dropped message for client %d: buffer full\n", client.id)
		}
	}
}

// handleClientWrite writes messages to the client
func (s *KVServer) handleClientWrite(client *Client) {
	for {
		select {
		case message := <-client.messageChan:
			_, err := client.conn.Write(message)
			if err != nil {
				fmt.Printf("Error writing to client %d: %v\n", client.id, err)
				s.removeClientChan <- client.id
				return
			}
		case <-client.done:
			return
		}
	}
}

func (s *KVServer) CountActive() int {
	respChan := make(chan int)
	s.countChan <- respChan
	return <-respChan
}

func (s *KVServer) CountDropped() int {
	respChan := make(chan int)
	s.droppedChan <- respChan
	return <-respChan
}

func (s *KVServer) Close() {
	if !s.started || s.closed {
		return
	}

	s.closed = true
    s.started = false

	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}
}
