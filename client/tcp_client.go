package tris

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/fvbock/tris/util"
	"net"
)

/*
Client
*/
type TCPClient struct {
	Dsn       *DSN
	Conn      net.Conn
	connected bool
	ActiveDb  string
	SessionId string
}

func NewTCPClient(dsn *DSN) (c *TCPClient, err error) {
	c = &TCPClient{
		Dsn:       dsn,
		connected: false,
	}
	return
}

/*
Set up the connection to a goxgo service specified by the DSN
*/
func (c *TCPClient) Dial() (err error) {
	if c.connected {
		err = errors.New("Already connected")
		return
	}
	c.Conn, err = net.Dial(c.Dsn.Protocol, fmt.Sprintf("%s:%v", c.Dsn.Host, c.Dsn.Port))
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot open connection:%v\n", err))
		return
	}
	c.connected = true
	return
}

/*
Close sends the EXIT commands and then closes the clients zmq socket
*/
func (c *TCPClient) Close() {
	_, _ = c.Send("EXIT")
	if c.connected {
		c.Conn.Close()
	}
	c.connected = false
	return
}

/*
Serialize the given payload, send it over the wire and return the
response data
*/
func (c *TCPClient) Send(msg string) (response []byte, err error) {
	// send the message
	_, err = c.Conn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Println("W", err)
	}

	// read the reply
	var rLength int64
	r := bufio.NewReader(c.Conn)
	rLength, prefLength, err := tris.ReadVarint(r)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("prefLength:", prefLength, "response length:", rLength)
	for n := 0; n < 4-prefLength; n++ {
		// fmt.Println(">")
		_, err = r.ReadByte()
		if err != nil {
			fmt.Println(err)
		}
	}

	buf := make([]byte, rLength)
	var responseBuffer bytes.Buffer
	for {
		bytesRead, err := r.Read(buf)
		if err != nil {
			fmt.Println("error reading response buf:", err)
			break
		}
		rLength -= int64(bytesRead)
		responseBuffer.Write(buf)
		// fmt.Println(bytesRead, "*", rLength)
		if rLength < 1 {
			// fmt.Println("rLength < 1")
			break
		}
	}
	response = responseBuffer.Bytes()
	return
}
