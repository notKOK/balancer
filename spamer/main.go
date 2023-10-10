package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	requestURL := "http://localhost:8000"
	fmt.Println(requestURL)
	go spam(requestURL)
	time.Sleep(time.Minute * 20)
}

func spam(requestURL string) {
	t := time.NewTicker(time.Millisecond)
	for {
		select {
		case <-t.C:
			go func() {
				res, err := http.Get(requestURL)
				log.Printf("response status: %d\n", res.StatusCode)
				if err != nil {
					fmt.Printf("error making http request: %s\n", err)
					os.Exit(1)
				}
			}()
		}
	}
}
