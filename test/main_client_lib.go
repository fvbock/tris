package main

import (
	// "fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/fvbock/tris/client"
	// "github.com/fvbock/tris/server"
	// "log"
	"runtime"
	// "sync"
	// "time"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func main() {
	dsn := &tris.DSN{
		Protocol: "tcp",
		Host:     "127.0.0.1",
		Port:     6000,
	}
	ctx, _ := zmq.NewContext()

	// todo: use a pool
	// client, err := pool.Get()
	client, _ := tris.NewClient(dsn, ctx)
	client.Dial()
	// todo: defer pool.Put(client)
	defer client.Close()

	client.Ping()
	client.Select("foo")
	client.DbInfo()

	client.Has("foo")
	client.Has("food")

	client.HasCount("foo")
	client.HasCount("food")

	client.HasPrefix("foo")
	client.HasPrefix("food")

	client.Members()
	client.PrefixMembers("foo")

	client.Add("food")
	client.Add("food")
	client.Del("food")

	// // check conn
	// r, err := client.Send("PING")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// response := tris.Unserialize(r)
	// if response.ReturnCode != tris.COMMAND_OK {
	// 	fmt.Printf("Initial PING failed:\n%v\n", response)
	// } else {
	// 	log.Println(response)
	// }

}

// func main() {
// 	// test calls
// 	var nrq int = 1000
// 	// var responses [][]byte
// 	var responses []string
// 	ctx, _ := zmq.NewContext()
// 	dsn := &tris.DSN{
// 		Protocol: "tcp",
// 		// Host:     "dogma",
// 		Host: "127.0.0.1",
// 		Port: 6000,
// 	}
// 	wg := sync.WaitGroup{}
// 	startA := time.Now()
// 	for i := 0; i < nrq; i++ {
// 		// if i%4 != 0 {
// 		time.Sleep(time.Millisecond * time.Duration(1000/nrq))
// 		// }
// 		wg.Add(1)
// 		go func(msgnr int, ztx *zmq.Context) {
// 			// start := time.Now()
// 			client, err := tris.NewClient(dsn, ctx)
// 			client.Dial()
// 			// defer client.Close()

// 			start := time.Now()
// 			msg := fmt.Sprintf("INFO\n")
// 			r, err := client.Send(msg)

// 			if err != nil {
// 				log.Printf("Call failed: %v\n", err)
// 			}
// 			// log.Println("GOT a reply:", string(r))
// 			responses = append(responses, string(r))
// 			log.Printf("done in %v\n", time.Since(start))
// 			client.Close()
// 			wg.Done()
// 		}(i, ctx)
// 	}
// 	wg.Wait()
// 	log.Printf("%v reqs done in %v\n", nrq, time.Since(startA))
// 	// log.Println("responses:", responses)
// 	// for _, r := range responses {
// 	// 	log.Println("response:", r)
// 	// }
// 	time.Sleep(1 * time.Second)
// 	log.Println("exit")
// }
