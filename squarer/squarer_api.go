package squarer

type Squarer interface {
	Initialize(chan int) <-chan int
	Close()
}
