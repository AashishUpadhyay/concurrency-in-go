package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var receivedOrdersCh = make(chan order)
	var validOrdersCh = make(chan order)
	var invalidOrdersCh = make(chan invalidorder)
	wg.Add(1)
	go receiveOrders(receivedOrdersCh)
	go validateOrders(receivedOrdersCh, validOrdersCh, invalidOrdersCh)
	go func(vo <-chan order, ivo <-chan invalidorder) {
	loop:
		for {
			select {
			case order, ok := <-vo:
				if ok {
					fmt.Printf("Valid order received : %v\n", order)
				} else {
					break loop
				}
			case order, ok := <-ivo:
				if ok {
					fmt.Printf("Invalid order received : %v. Issue : %v\n", order.order, order.err)
				} else {
					break loop
				}
			}
		}
		wg.Done()
	}(validOrdersCh, invalidOrdersCh)
	wg.Wait()
	fmt.Println(orders)
}

func receiveOrders(out chan<- order) {
	for _, rawOrder := range rawOrders {
		var newOrder order
		err := json.Unmarshal([]byte(rawOrder), &newOrder)
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Printf("Received order : %d\n", newOrder.ProductCode)
		out <- newOrder
	}
	close(out)
}

func validateOrders(in <-chan order, out chan<- order, errChan chan<- invalidorder) {
	for order := range in {
		fmt.Printf("Validating order : %d\n", order.ProductCode)
		if order.Quantity <= 0 {
			// error condition
			errChan <- invalidorder{order: order, err: errors.New("Quantity must be greater than zero")}
		} else {
			out <- order
		}
	}
	close(out)
	close(errChan)
}

var rawOrders = []string{
	`{"productCode":1111, "quantity":-5, "status":1}`,
	`{"productCode":2222, "quantity":42.3, "status":1}`,
	`{"productCode":3333, "quantity":19, "status":1}`,
}
