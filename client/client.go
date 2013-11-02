package tris

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

const (
	VERSION = "0.0.2"
)

/*
Structure containing the connection endpoint
*/
type DSN struct {
	Protocol string
	Host     string
	Port     int
}

/*
Client
*/
type Client struct {
	Dsn       *DSN
	Context   *zmq.Context
	Socket    *zmq.Socket
	connected bool
	ActiveDb  string
	SessionId string
}

func NewClient(dsn *DSN, ctx *zmq.Context) (c *Client, err error) {
	c = &Client{
		Dsn:       dsn,
		connected: false,
	}
	if ctx != nil {
		c.Context = ctx
	} else {
		c.Context, err = zmq.NewContext()
		if err != nil {
			err = errors.New(fmt.Sprintf("Cannot initialize Context:%v\n", err))
			return
		}
	}
	return
}

/*
Set up the connection to a goxgo service specified by the DSN
*/
func (c *Client) Dial() (err error) {
	if c.connected {
		err = errors.New("Already connected")
		return
	}
	c.Socket, err = c.Context.NewSocket(zmq.REQ)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot open Socket:%v\n", err))
		return
	}
	c.Socket.SetSockOptInt(zmq.LINGER, 0)
	c.Socket.Connect(fmt.Sprintf("%v://%v:%v", c.Dsn.Protocol, c.Dsn.Host, c.Dsn.Port))
	c.connected = true
	return
}

/*
Close sends the EXIT commands and then closes the clients zmq socket
*/
func (c *Client) Close() {
	_, _ = c.Send("EXIT")
	if c.connected {
		c.Socket.Close()
	}
	c.connected = false
	return
}

/*
Serialize the given payload, send it over the wire and return the
response data
*/
func (c *Client) Send(msg string) (response []byte, err error) {
	c.Socket.Send([]byte(msg+"\n"), 0)
	response, err = c.Socket.Recv(0)
	return
}
