package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOrderManagement(t *testing.T) {
	// Since orderManagement() doesn't return anything, we're mainly testing it doesn't panic
	orderManagement()
}

func TestMutexExample(t *testing.T) {
	mutexExample()
	// Add actual assertions if you want to test for race conditions
}

func TestSyncOnceExample(t *testing.T) {
	// Test that sync.Once executes only once
	syncOnceExample()
	syncOnceExample() // Should not open connection second time
}

func TestContextExample(t *testing.T) {
	done := make(chan bool)
	go func() {
		contextExample()
		done <- true
	}()

	select {
	case <-done:
		// Test passed
	case <-time.After(3 * time.Second):
		t.Error("contextExample didn't complete within expected timeframe")
	}
}

func TestValidateOrders(t *testing.T) {
	// Create test input channel
	inChan := make(chan order)
	validChan, invalidChan := validateOrders(inChan)

	go func() {
		// Send test orders
		inChan <- order{Quantity: 1}  // valid
		inChan <- order{Quantity: 0}  // invalid
		inChan <- order{Quantity: -1} // invalid
		close(inChan)
	}()

	validCount := 0
	invalidCount := 0

	// Start goroutine to read valid orders
	done := make(chan bool)
	go func() {
		for range validChan {
			validCount++
		}
		done <- true
	}()

	// Read invalid orders
	for range invalidChan {
		invalidCount++
	}

	<-done // Wait for valid channel to be done

	if validCount != 1 {
		t.Errorf("Expected 1 valid order, got %d", validCount)
	}
	if invalidCount != 2 {
		t.Errorf("Expected 2 invalid orders, got %d", invalidCount)
	}
}

func TestReserveOrders(t *testing.T) {
	inChan := make(chan order)
	outChan := reserveOrders(inChan)

	go func() {
		// Send test orders
		inChan <- order{Status: received}
		inChan <- order{Status: received}
		close(inChan)
	}()

	count := 0
	for order := range outChan {
		if order.Status != reserved {
			t.Errorf("Expected status %d, got %d", reserved, order.Status)
		}
		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 orders, got %d", count)
	}
}

func TestReceiveOrdersDuration(t *testing.T) {
	out := make(chan int)
	duration := make(chan time.Duration) // Buffered channel to avoid blocking

	go func() {
		startGet := time.Now()
		// Simulate some work
		time.Sleep(500 * time.Millisecond)
		out <- 1
		duration <- time.Since(startGet) // Send duration without blocking
		close(out)
		close(duration) // Close the channel after sending
	}()

	fmt.Printf("Order value %d", <-out)

	// Receive and print the duration
	for dur := range duration {
		fmt.Println("Duration:", dur)
	}
}

func TestReceiveOrders(t *testing.T) {
	out := receiveOrders()
	count := 0
	for out != nil {
		select {
		case order, ok := <-out:
			if ok {
				if order.Status != received {
					t.Errorf("Expected status %d, got %d", received, order.Status)
				}
				count++
			} else {
				out = nil // Set to nil when the channel is closed
			}
		}
	}
	if count != 4 {
		t.Errorf("Expected 4 orders, got %d", count)
	}
}

func TestKVControllerLoadTest(t *testing.T) {
	concurrentUsers := 1
	requestsPerUser := 1

	var wg sync.WaitGroup

	successChan := make(chan time.Duration, concurrentUsers*requestsPerUser)
	wg.Add(concurrentUsers)
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			defer wg.Done() // Ensure that Done is called when goroutine finishes
			fmt.Printf("User %d started\n", userID)
			for j := 0; j < requestsPerUser; j++ {
				startPost := time.Now()
				postDuration := time.Since(startPost)
				successChan <- postDuration
				startGet := time.Now()
				time.Sleep(100 * time.Millisecond) // Placeholder for actual request time
				getDuration := time.Since(startGet)
				successChan <- getDuration
			}
		}(i)
	}

	wg.Wait()
	close(successChan)
	var totalDuration time.Duration
	successCount := 0

	for duration := range successChan {
		totalDuration += duration
		successCount++
	}
}
