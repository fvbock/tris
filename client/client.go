package tris

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

/*
Structure containing the connection endpoint

For now the only supported protocol option is "tcp" which will be a ZMQ
connection.

TODO: http protocol support. will need some refactoring
*/
type DSN struct {
	Protocol string
	Host     string
	Port     int
}

/*
Conn - ZMQ connection structure
*/
type Conn struct {
	Dsn       *DSN
	Context   *zmq.Context
	Socket    *zmq.Socket
	connected bool
}

/*
Set up the connection to a goxgo service specified by the DSN
*/
func (c *Conn) Dial(dsn *DSN) (err error) {
	c.Context, err = zmq.NewContext()
	if err != nil {
		return err
	}

	c.Socket, err = c.Context.NewSocket(zmq.REQ)
	if err != nil {
		return err
	}

	// TODO: add a conn/conf parameter to set a timeout
	c.Socket.SetSockOptInt(zmq.LINGER, 0)
	if err == nil {
		c.connected = true
	}
	c.Socket.Connect(fmt.Sprintf("%v://%v:%v", dsn.Protocol, dsn.Host, dsn.Port))
	return
}

/*
Close the connections zmq socket and zmq context
*/
func (c *Conn) Close() {
	if c.connected {
		c.Socket.Close()
		// fmt.Println("socket closed...")
	}
	c.connected = false
	return
}

/*
Serialize the given payload, send it over the wire and return the
response data
*/
func (c *Conn) Send(msg string) (response []byte, err error) {
	c.Socket.Send([]byte(msg), 0)
	response, _ = c.Socket.Recv(0)
	return
}

/*
Convinience function that will create a connection,
send a payload and Unserialize the reponse into a response
structure.
*/
func Call(dsn *DSN, msg string) (response []byte, err error) {
	c := Conn{Dsn: dsn}
	err = c.Dial(dsn)
	if err != nil {
		err = errors.New(fmt.Sprintf("Shit hit the fan: %v.", err))
		return
	}
	defer c.Close()
	response, err = c.Send(msg)
	if err != nil {
		fmt.Println("Error sending msg:", err)
		return
	}
	return
}
