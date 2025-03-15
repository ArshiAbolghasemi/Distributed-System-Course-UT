package client

type KVClient interface {
	Connect(address string) error
	SendMessage(message string) error
	ReadResponse() (string, error)
	Close()
}
