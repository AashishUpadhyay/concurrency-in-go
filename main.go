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
	go receiveOrders(receivedOrdersCh)
	go validateOrders(receivedOrdersCh, validOrdersCh, invalidOrdersCh)
	wg.Add(1)
	go func() {
		order := <-validOrdersCh
		fmt.Printf("Valid order received : %v\n", order)
		wg.Done()
	}()
	go func() {
		order := <-invalidOrdersCh
		fmt.Printf("Inaalid order received : %v. Issue : %v\n", order.order, order.err)
		wg.Done()
	}()
	wg.Wait()
	fmt.Println(orders)
}

func receiveOrders(out chan order) {
	for _, rawOrder := range rawOrders {
		var newOrder order
		err := json.Unmarshal([]byte(rawOrder), &newOrder)
		if err != nil {
			log.Print(err)
			continue
		}
		out <- newOrder
	}
}

func validateOrders(in chan order, out chan order, errChan chan invalidorder) {
	order := <-in
	if order.Quantity <= 0 {
		// error condition
		errChan <- invalidorder{order: order, err: errors.New("Quantity must be greater than zero")}
	} else {
		out <- order
	}
}

var rawOrders = []string{
	`{"productCode":1111, "quantity":-5, "status":1}`,
	`{"productCode":2222, "quantity":42.3, "status":1}`,
	`{"productCode":3333, "quantity":19, "status":1}`,
}
