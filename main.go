package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	orderManagement()
	mutexExample()
	syncOnceExample()
	syncOnceExample()
	contextExample()
}

func contextExample() {
	// cancel after a while example
	var ctx, cancel = context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for range time.Tick(500 * time.Millisecond) {
			if ctx.Err() != nil {
				log.Println(ctx.Err())
				return
			}
			fmt.Println("tick!")
		}
	}(ctx)

	time.Sleep(2 * time.Second)
	cancel()
	wg.Wait()
}

var m sync.Once

func syncOnceExample() {
	m.Do(func() {
		db, err := sql.Open("sqlite3", "./foo.db")
		fmt.Print("Opening Connnection!\n")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	})
}

func mutexExample() {
	// run using go run -race . and detect race conditions
	var wg sync.WaitGroup
	// var m sync.Mutex
	var arr = []int{}

	const iterations = 1000
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			// m.Lock()
			// defer m.Unlock()
			arr = append(arr, 1)
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Printf("Length : %d\n", len(arr))
}

func orderManagement() {
	var wg sync.WaitGroup
	receivedOrdersCh := receiveOrders()
	validOrdersCh, invalidOrdersCh := validateOrders(receivedOrdersCh)
	reserveOrdersCh := reserveOrders(validOrdersCh)
	wg.Add(1)
	go func(ivo <-chan invalidorder) {
		for invorder := range ivo {
			fmt.Printf("Invalid order : %v\n", invorder.order.ProductCode)
		}
		wg.Done()
	}(invalidOrdersCh)

	const workers = 3
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(ro <-chan order) {
			for rorder := range ro {
				fmt.Printf("Order filled : %v\n", rorder)
			}
			wg.Done()
		}(reserveOrdersCh)
	}
	wg.Wait()
}

func receiveOrders() <-chan order {
	var out = make(chan order)
	go func() {
		for _, rawOrder := range rawOrders {
			var newOrder order
			err := json.Unmarshal([]byte(rawOrder), &newOrder)
			if err != nil {
				log.Print(err)
				continue
			}
			newOrder.Status = received
			out <- newOrder
		}
		close(out)
	}()
	return out
}

func validateOrders(in <-chan order) (<-chan order, <-chan invalidorder) {
	var out = make(chan order)
	var errChan = make(chan invalidorder)
	go func() {
		for order := range in {
			if order.Quantity <= 0 {
				// error condition
				errChan <- invalidorder{order: order, err: errors.New("quantity must be greater than zero")}
			} else {
				out <- order
			}
		}
		close(out)
		close(errChan)
	}()
	return out, errChan
}

func reserveOrders(in <-chan order) <-chan order {
	var out = make(chan order)
	var wg sync.WaitGroup
	const workers = 3
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for order := range in {
				order.Status = reserved
				out <- order
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

var rawOrders = []string{
	`{"productCode":1111, "quantity":5, "status":1}`,
	`{"productCode":2222, "quantity":42.3, "status":1}`,
	`{"productCode":3333, "quantity":19, "status":1}`,
	`{"productCode":444, "quantity":-19, "status":1}`,
}
