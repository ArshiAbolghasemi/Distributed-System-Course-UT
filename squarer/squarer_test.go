package squarer

import (
	"fmt"
	"testing"
	"time"
)

const (
	timeoutMillis = 5000
)

func TestBasicCorrectness(t *testing.T) {
	fmt.Println("Running TestBasicCorrectness.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	go func() {
		input <- 2
	}()
	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		t.Error("Test timed out.")
	case result := <-squares:
		if result != 4 {
			t.Error("Error, got result", result, ", expected 4 (=2^2).")
		}
	}
}

func TestSequentialNumbers(t *testing.T) {
	fmt.Println("Running TestSequentialNumbers.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	
	go func() {
		input <- 1
		input <- 3
		input <- 5
		input <- 10
	}()
	
	expectedResults := []int{1, 9, 25, 100}
	for _, expected := range expectedResults {
		timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
		select {
		case <-timeoutChan:
			t.Error("Test timed out waiting for result.")
			return
		case result := <-squares:
			if result != expected {
				t.Errorf("Error, got result %d, expected %d", result, expected)
				return
			}
		}
	}
}

func TestConcurrentProcessing(t *testing.T) {
	fmt.Println("Running TestConcurrentProcessing.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	
	count := 10
	
	received := make(chan int, count)
	
	go func() {
		for range count {
			select {
			case val := <-squares:
				received <- val
			case <-time.After(time.Duration(timeoutMillis) * time.Millisecond):
				return
			}
		}
	}()

    for i := 1; i <= count; i++ {
        go func(val int) {
            input <- i
        }(i)
    }
	
	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	
	receivedValues := make(map[int]bool)
    for i := 1; i <= count; i++ {
		select {
		case <-timeoutChan:
			t.Errorf("Test timed out after receiving %d values.", i)
			return
		case val := <-received:
			receivedValues[val] = true
		}
	}
	
	for i := 1; i <= count; i++ {
		square := i * i
		if !receivedValues[square] {
			t.Errorf("Expected square %d (=%d^2) was not received.", square, i)
		}
	}	
}
