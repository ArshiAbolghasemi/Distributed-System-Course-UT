package squarer

type SquarerImpl struct {
	input  chan int
	output chan int
	close  chan bool
	closed chan bool
}

func (sq *SquarerImpl) Initialize(input chan int) <-chan int {
	sq.input = input
	sq.close = make(chan bool)
	sq.closed = make(chan bool)
	sq.output = make(chan int)
	go sq.work()
	return sq.output
}

func (sq *SquarerImpl) Close() {
	sq.close <- true
	<-sq.closed
}

func (sq *SquarerImpl) work() {
	var toPush int
	dummy := make(chan int)
	pushOn := dummy
	pullOn := sq.input
	for {
		select {
		case unsquared := <-pullOn:
			toPush = unsquared * unsquared
			pushOn = sq.output
			pullOn = nil
		case pushOn <- toPush:
			pushOn = dummy
			pullOn = sq.input
		case <-sq.close:
			sq.closed <- true
			return
		}
	}
}
