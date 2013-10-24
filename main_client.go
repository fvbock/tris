package main

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/fvbock/tris/client"
	"log"
	"runtime"
	"sync"
	"time"
)

func init() {
	runtime.GOMAXPROCS(2)
}

func main() {
	// test calls
	var nrq int = 500
	// var responses [][]byte
	var responses []string
	ctx, _ := zmq.NewContext()
	dsn := &tris.DSN{
		Protocol: "tcp",
		Host:     "localhost",
		Port:     6000,
	}
	wg := sync.WaitGroup{}
	startA := time.Now()
	for i := 0; i < nrq; i++ {
		if i%4 != 0 {
			time.Sleep(time.Millisecond * time.Duration(1000/nrq))
		}
		wg.Add(1)
		go func(msgnr int, ztx *zmq.Context) {
			// start := time.Now()
			sock, _ := ztx.NewSocket(zmq.REQ)
			sock.Connect(fmt.Sprintf("%v://%v:%v", dsn.Protocol, dsn.Host, dsn.Port))
			// log.Printf("connect done in %v\n", time.Since(start))
			start := time.Now()
			msg := fmt.Sprintf("INFO\nSELECT a\nINFO\n")
			sock.Send([]byte(msg), 0)
			r, err := sock.Recv(0)
			if err != nil {
				log.Printf("Call failed: %v\n", err)
			}
			// log.Println("GOT a reply:", string(r))
			responses = append(responses, string(r))
			log.Printf("done in %v\n", time.Since(start))
			wg.Done()
		}(i, ctx)
	}
	wg.Wait()
	log.Printf("%v reqs done in %v\n", nrq, time.Since(startA))
	// log.Println("responses:", responses)
	// for _, r := range responses {
	// 	log.Println("response:", r)
	// }
}
