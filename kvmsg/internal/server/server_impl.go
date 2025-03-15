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
	conn    net.Conn
	id      int
	inChan  chan Command
	outChan chan []byte
	done    chan struct{}
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

func New() *KVServer {
	return NewWithStore(kvstore.New())
}

func NewWithStore(store kvstore.KVStore) *KVServer {
	return &KVServer{
		clients:            make(map[int]*Client),
		nextClientID:       1,
		kvstore:            store,
		opsChan:            make(chan Command),
		registerClientChan: make(chan *Client),
		removeClientChan:   make(chan int),
		countChan:          make(chan chan int),
		droppedChan:        make(chan chan int),
		shutdown:           make(chan struct{}),
	}
}

func (s *KVServer) Start(port int) error {
	if s.started {
		return fmt.Errorf("server already started")
	}

	addr, err := config.GetServerAddress()
	if err != nil {
		return err
	}
	protocol, err := config.GetServerProtocol()
	if err != nil {
		return err
	}
	listener, err := net.Listen(protocol, addr)
	if err != nil {
		return err
	}

	s.listener = listener
	s.port = port
	s.started = true
	s.closed = false

	go s.handleKVStore()

	go s.manageClients()

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
			if client, exists := s.clients[clientID]; exists {
				delete(s.clients, clientID)
				s.activeCount--
				s.droppedCount++

				close(client.outChan)
				close(client.done)
			}

		case respChan := <-s.countChan:
			respChan <- s.activeCount

		case respChan := <-s.droppedChan:
			respChan <- s.droppedCount

		case <-s.shutdown:
			for _, client := range s.clients {
				close(client.outChan)
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

				s.notifyClients(fmt.Sprintf("NOTIFY PUT %s", cmd.key))

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

				s.notifyClients(fmt.Sprintf("NOTIFY DELETE %s", cmd.key))

			case OpUpdate:
				s.kvstore.Update(cmd.key, cmd.oldValue, cmd.value)
				cmd.respChan <- []byte("OK")

				s.notifyClients(fmt.Sprintf("NOTIFY UPDATE %s", cmd.key))
			}

		case <-s.shutdown:
			return
		}
	}
}

func (s *KVServer) notifyClients(notification string) {
	for id, client := range s.clients {
		select {
		case client.outChan <- []byte(notification + "\n"):
		default:
			fmt.Printf("Failed to notify msg %s to client %d: buffer full\n", notification, id)
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

		chanBufLimit, err := config.GetClientChanBufLimit()
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
			conn:    conn,
			id:      s.nextClientID,
			inChan:  make(chan Command),
			outChan: make(chan []byte, chanBufLimit),
			done:    make(chan struct{}),
		}
		s.nextClientID++

		s.registerClientChan <- client

		go s.handleClientRead(client)
		go s.handleClientWrite(client)
		go s.processClientCommands(client)
	}
}

func (s *KVServer) handleClientRead(client *Client) {
	reader := bufio.NewReader(client.conn)

	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from client %d: %v\n", client.id, err)
			}
			s.removeClientChan <- client.id
			return
		}

		cmd, err := ParseCommand(message)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v\n", err)
			select {
			case client.outChan <- []byte(errMsg):
			default:
				fmt.Printf("Dropped error message for client %d: buffer full\n", client.id)
			}
			continue
		}

		select {
		case client.inChan <- cmd:
		case <-client.done:
			return
		}
	}
}

func (s *KVServer) processClientCommands(client *Client) {
	for {
		select {
		case cmd := <-client.inChan:
			s.opsChan <- cmd

			resp := <-cmd.respChan
			resp = append(resp, '\n')

			select {
			case client.outChan <- resp:
			default:
				fmt.Printf("Dropped response for client %d: buffer full\n", client.id)
			}

		case <-client.done:
			return
		}
	}
}

func (s *KVServer) handleClientWrite(client *Client) {
	for {
		select {
		case message, ok := <-client.outChan:
			if !ok {
				return
			}

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
