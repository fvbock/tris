package main

import (
	"fmt"
	"github.com/fvbock/tris/client"
	"log"
	"runtime"
	"sync"
	"time"
)

func init() {
	runtime.GOMAXPROCS(1)
}

func main() {
	// test calls
	var nrq int = 1000
	// var responses [][]byte
	var responses []string
	dsn := &tris.DSN{
		Protocol: "tcp",
		Host:     "localhost",
		Port:     16000,
	}
	wg := sync.WaitGroup{}
	startA := time.Now()
	for i := 0; i < nrq; i++ {
		// if i%2 != 0 {
		// 	time.Sleep(time.Millisecond * time.Duration(2000/nrq))
		// }
		time.Sleep(time.Millisecond * time.Duration(2000/nrq))
		wg.Add(1)
		go func(msgnr int) {
			// start := time.Now()
			c, _ := tris.NewTCPClient(dsn)
			err := c.Dial()
			for err != nil {
				log.Printf("Dial failed: %v\n", err)
				err = c.Dial()
			}
			// log.Printf("connect done in %v\n", time.Since(start))
			start := time.Now()
			msg := fmt.Sprintf("INFO\nSELECT a\nINFO")
			r, err := c.Send(msg)
			if err != nil {
				log.Printf("Call failed: %v\n", err)
			}
			// log.Println("GOT a reply:", string(r))
			responses = append(responses, string(r))
			log.Printf("done in %v\n", time.Since(start))
			// exit
			c.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()
	log.Printf("%v reqs done in %v\n", nrq, time.Since(startA))
	// log.Println("responses:", responses)
	// for _, r := range responses {
	// 	log.Println("response:", r)
	// }
}
