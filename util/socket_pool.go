package util

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

type ZmqSocketPool struct {
	Ctx        *zmq.Context
	Pool       chan *zmq.Socket
	SocketType zmq.SocketType
	PoolSize   int
}

func InitializeZmqSocketPool(sType zmq.SocketType, connCount int) (p *ZmqSocketPool, err error) {

}

func (p *ZmqSocketPool) Get() (s *zmq.Socket, err error) {

}

func (p *ZmqSocketPool) Release() (err error) {

}
