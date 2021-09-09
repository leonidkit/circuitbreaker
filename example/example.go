package main

import (
	"circuit_breaker"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

const (
	PROBABILITY_OF_SERVER_FAILURE = 0.05
	REQUESTS_NUM                  = 1000
	GOROUTINES_NUM                = 10
)

func setUpTestServer() *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rand.Float32() < PROBABILITY_OF_SERVER_FAILURE {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
	return httptest.NewServer(h)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// setup server
	srv := setUpTestServer()

	// CircuitBreaker initialization
	cb := circuit_breaker.New(circuit_breaker.Settings{
		Timeout:     1 * time.Millisecond,
		Threshold:   2,
		MaxRequests: 10,
	})

	// define routine
	routine := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < REQUESTS_NUM; i++ {
			time.Sleep(10 * time.Millisecond)
			// check if we can make a request
			if !cb.Allow() {
				continue
			}
			resp, err := http.Get(srv.URL)
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode != 200 {
				// register error
				cb.RegisterError()
			} else {
				// register ok
				cb.RegisterOK()
			}
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(GOROUTINES_NUM)
	fmt.Printf("Started %d goroutines for %d requests. Waiting...\n", GOROUTINES_NUM, REQUESTS_NUM*GOROUTINES_NUM)
	for i := 0; i < GOROUTINES_NUM; i++ {
		go routine(wg)
	}
	wg.Wait()
	fmt.Printf("%+v\n", cb.Counters())
}
