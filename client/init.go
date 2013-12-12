package tris

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

var (
	TrisClientZmqContext *zmq.Context
)

func init() {
	var err error
	TrisClientZmqContext, err = zmq.NewContext()
	if err != nil {
		panic(fmt.Sprintf("Cannot initialize Tris ZMQ Context:\n%v\nBailing.", err))
	}
	fmt.Println("TrisClientZmqContext initialized.")
}
