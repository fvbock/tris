package tris

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/fvbock/tris/server"
	"log"
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

// func (c *Client) exec(cmd tris.Command, args ...string) {
func (c *Client) exec(cmd tris.Command, args ...string) (response *tris.Reply, err error) {
	msg := cmd.Name()
	for _, arg := range args {
		msg += " " + arg
	}
	r, err := c.Send(msg)
	if err != nil {
		fmt.Println("Error:", err)
	}
	response = tris.Unserialize(r)
	if response.ReturnCode != tris.COMMAND_OK {
		log.Printf("FAILED:\ncmd: %s\nargs: %v\nresponse: %v\n", cmd, args, response)
	} else {
		log.Printf("cmd: %s\nargs: %v\nresponse: %v\n", cmd.Name(), args, response)
	}
	response.Print()
	return
}

func (c *Client) Ping() {
	c.exec(&tris.CommandPing{})
}

func (c *Client) Select(dbname string) {
	c.exec(&tris.CommandSelect{}, dbname)
}

func (c *Client) DbInfo() {
	c.exec(&tris.CommandDbInfo{})
}

func (c *Client) Info() {
	c.exec(&tris.CommandInfo{})
}

// TrisCommands = append(TrisCommands, &CommandExit{})

func (c *Client) Save() {
	c.exec(&tris.CommandSave{})
}

// TrisCommands = append(TrisCommands, &CommandSave{})
// TrisCommands = append(TrisCommands, &CommandImportDb{})
// TrisCommands = append(TrisCommands, &CommandMergeDb{})

func (c *Client) Create(dbname string) {
	c.exec(&tris.CommandCreateTrie{}, dbname)
}

func (c *Client) Add(key string) {
	c.exec(&tris.CommandAdd{}, key)
}

func (c *Client) Del(key string) {
	c.exec(&tris.CommandDel{}, key)
}

func (c *Client) Has(key string) {
	c.exec(&tris.CommandHas{}, key)
}

func (c *Client) HasCount(key string) {
	c.exec(&tris.CommandHasCount{}, key)
}

func (c *Client) HasPrefix(key string) {
	c.exec(&tris.CommandHasPrefix{}, key)
}

func (c *Client) Members() {
	c.exec(&tris.CommandMembers{})
}

func (c *Client) PrefixMembers(key string) {
	c.exec(&tris.CommandPrefixMembers{}, key)
}

// TrisCommands = append(TrisCommands, &CommandTree{})
// TrisCommands = append(TrisCommands, &CommandTiming{})
// TrisCommands = append(TrisCommands, &CommandShutdown{})
