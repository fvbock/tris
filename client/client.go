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

// type ClientCommand interface{

// }

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
	// Commands map[string]ClientCommand
}

func NewClient(dsn *DSN) (c *Client, err error) {
	c = &Client{
		Dsn:       dsn,
		Context:   TrisClientZmqContext,
		connected: false,
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
		// log.Printf("cmd: %s\nargs: %v\nresponse: %v\n", cmd.Name(), args, response)
	}
	// response.Print()
	return
}

func (c *Client) Ping() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandPing{})
	if r.ReturnCode != tris.COMMAND_OK || err != nil {
		// ???
	}
	return
}

func (c *Client) Select(dbname string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandSelect{}, dbname)
	if r.ReturnCode != tris.COMMAND_OK || err != nil {
		// TODO
	}
	c.ActiveDb = dbname
	return
}

func (c *Client) DbInfo() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandDbInfo{})
	return
}

func (c *Client) Info() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandInfo{})
	return
}

// TrisCommands = append(TrisCommands, &CommandExit{})

func (c *Client) Save() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandSave{})
	return
}

func (c *Client) ImportDb(fname string, dbname string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandImportDb{}, fname, dbname)
	return
}

func (c *Client) MergeDb(fname string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandMergeDb{}, fname)
	return
}

func (c *Client) Create(dbname string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandCreateTrie{}, dbname)
	return
}

func (c *Client) Add(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandAdd{}, key)
	return
}

func (c *Client) Del(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandDel{}, key)
	return
}

func (c *Client) Has(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandHas{}, key)
	return
}

func (c *Client) HasCount(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandHasCount{}, key)
	return
}

func (c *Client) HasPrefix(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandHasPrefix{}, key)
	return
}

func (c *Client) Members() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandMembers{})
	return
}

func (c *Client) PrefixMembers(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandPrefixMembers{}, key)
	return
}

func (c *Client) Tree() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandTree{})
	return
}

func (c *Client) Timing() (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandTiming{})
	return
}

// TrisCommands = append(TrisCommands, &CommandShutdown{})

func (c *Client) Help(key string) (r *tris.Reply, err error) {
	r, err = c.exec(&tris.CommandHelp{}, key)
	return
}
