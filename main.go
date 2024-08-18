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
	wg.Add(1)
	receivedOrdersCh := receiveOrders()
	validOrdersCh, invalidOrdersCh := validateOrders(receivedOrdersCh)
	reserveOrdersCh := reserveOrders(validOrdersCh)
	filledOrdersCh := filledOrders(reserveOrdersCh)
	go func(vo <-chan order, ivo <-chan invalidorder) {
	loop:
		for {
			select {
			case order, ok := <-vo:
				if ok {
					fmt.Printf("Order filled : %v\n", order)
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
	}(filledOrdersCh, invalidOrdersCh)
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
				errChan <- invalidorder{order: order, err: errors.New("Quantity must be greater than zero")}
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
	go func() {
		for order := range in {
			if order.Status == received {
				order.Status = reserved
			}
			out <- order
		}
		close(out)
	}()
	return out
}

func filledOrders(in <-chan order) <-chan order {
	var out = make(chan order)
	go func() {
		for order := range in {
			if order.Status == reserved {
				order.Status = filled
			}
			out <- order
		}
		close(out)
	}()
	return out
}

var rawOrders = []string{
	`{"productCode":1111, "quantity":5, "status":1}`,
	`{"productCode":2222, "quantity":42.3, "status":1}`,
	`{"productCode":3333, "quantity":19, "status":1}`,
}
