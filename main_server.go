package main

import (
	"github.com/davecheney/profile"
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
		PersistOpsLimit:   100,
		PersistInterval:   300 * time.Second,
	}
}

func main() {
	profileConfig := profile.Config{
		CPUProfile:     true,
		MemProfile:     true,
		ProfilePath:    ".",  // store profiles in current directory
		NoShutdownHook: true, // do not hook SIGINT
	}
	prof := profile.Start(&profileConfig)
	defer prof.Stop()

	server, err := tris.NewServer(config)
	if err != nil {
		server.Log.Printf("Could not initialize server: %v\n", err)
	}
	server.Start() // Blocks until the server Stop()s

	server.Log.Println("Done")
}
