package main

import (
	"github.com/fvbock/tris/server"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	Log *log.Logger
)

func init() {
	runtime.GOMAXPROCS(2)
	Log = log.New(os.Stderr, "", log.LstdFlags)
}

func main() {
	server, err := tris.New(Log)
	server.RegisterCommands(tris.TrisCommands...)
	if err != nil {
		Log.Printf("Could not initialize server: %v\n", err)
	}
	server.Start()

	Log.Println("Wait for 10 sec")
	time.Sleep(1000 * time.Second)
	server.Stop()
	Log.Println("Done")
}
