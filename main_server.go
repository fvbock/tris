package main

import (
	"github.com/fvbock/tris/server"
	"runtime"
	"time"
)

var (
	config *tris.ServerConfig
)

func init() {
	runtime.GOMAXPROCS(4)
	config = &tris.ServerConfig{
		Protocol:          "tcp",
		Host:              "127.0.0.1",
		Port:              6000,
		DataDir:           "/home/morpheus/tris_data",
		StorageFilePrefix: "trie_",
	}
}

func main() {
	server, err := tris.NewServer(config)
	if err != nil {
		server.Log.Printf("Could not initialize server: %v\n", err)
	}
	server.Start()

	server.Log.Println("Wait for 10 min")
	time.Sleep(1000 * time.Second)
	server.Stop()
	server.Log.Println("Done")
}
