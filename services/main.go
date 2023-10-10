package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var mux sync.RWMutex
var requestNum atomic.Int32

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestNum.Add(1)
		w.WriteHeader(http.StatusOK)

		time.Sleep(time.Second)

		fmt.Fprintln(w, "Это тело ответа")

		defer requestNum.Add(-1)
	})

	go countRequest()

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		return
	}
}

func countRequest() {
	t := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-t.C:
			mux.RLock()
			log.Printf("number of requests: %d \n", requestNum.Load())
			mux.RUnlock()
		}
	}
}
