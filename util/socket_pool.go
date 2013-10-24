package util

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

type ZmqSocketPool struct {
	Ctx        *zmq.Context
	Pool       chan *zmq.Socket
	SocketType string
	PoolSize   int
}

func (p *ZmqSocketPool) Get() (s *zmq.Socket, err error) {

}

func (p *ZmqSocketPool) Release() (err error) {

}
