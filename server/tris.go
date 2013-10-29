package tris

import (
	"bytes"
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/fvbock/trie"
	"log"
	"strings"
	"sync"
	"time"
)

func init() {
	// CreateTrie := Command{"CREATE", trie.NewRefCountTrie}
}

type ServerConfig struct {
	Proto []string
	Host  string
	Port  []int

	DataDir string
}

const (
	DEFAULT_DB = "0"
)

type Client struct {
	Id           []byte
	Msg          []byte
	ActiveDbName string
	ActiveDb     *trie.RefCountTrie
	Cmds         []Command
	Args         [][]interface{}
	Response     []byte
}

func (c *Client) String() string {
	return fmt.Sprintf("Client ID: %v\nActive Db: %v\n", c.Id, c.ActiveDbName)
}

func NewClient(s *Server, id []byte) *Client {
	return &Client{
		Id:           id,
		ActiveDbName: DEFAULT_DB,
		ActiveDb:     s.Databases[DEFAULT_DB],
	}
}

type Command interface {
	Name() string
	Function(s *Server, c *Client, args ...interface{}) (reply *Reply)
	Flags() int
}

const (
	STATE_STOP    = 1
	STATE_STOPPED = 2
	STATE_RUNNING = 3

	COMMAND_FLAG_READ  = 1
	COMMAND_FLAG_WRITE = 2
	COMMAND_FLAG_ADMIN = 4

	COMMAND_REPLY_EMPTY  = 0
	COMMAND_REPLY_SINGLE = 1
	COMMAND_REPLY_MULTI  = 2

	COMMAND_OK   = 0
	COMMAND_FAIL = 1
)

type Server struct {
	sync.RWMutex
	Log               *log.Logger
	Config            *ServerConfig
	Commands          map[string]Command
	Databases         map[string]*trie.RefCountTrie
	DatabaseOpCount   map[string]int
	State             int
	Stateswitch       chan int
	CommandsProcessed int

	CycleLength      int64
	CheckStateChange time.Duration

	ActiveClients int

	// zeromq
	Host      string
	Port      int
	Context   *zmq.Context
	Socket    *zmq.Socket
	PollItems zmq.PollItems

	// CommandQueue  chan *Client
	ResponseQueue chan *Client
}

func New(logger *log.Logger) (s *Server, err error) {
	s = &Server{
		CommandsProcessed: 0,
		Log:               logger,
		Commands:          make(map[string]Command),
		Databases:         make(map[string]*trie.RefCountTrie),
		Stateswitch:       make(chan int, 1),
		ResponseQueue:     make(chan *Client),

		// CycleLength:      int64(time.Millisecond) * 1,
		CycleLength:      int64(time.Microsecond) * 500,
		CheckStateChange: time.Second * 1,
	}
	s.Initialize()
	return
}

func (s *Server) Initialize() {
	s.Databases[DEFAULT_DB] = trie.NewRefCountTrie()
}

func (s *Server) RegisterCommands(cmds ...Command) (err error) {
	for _, c := range cmds {
		if _, exists := s.Commands[c.Name()]; exists {
			err = errors.New(fmt.Sprintf("Command %s has already been registered.", c.Name()))
			return
		}
		s.Log.Println("Registering command", c.Name())
		s.Commands[c.Name()] = c
	}
	return
}

// Think of REQ and DEALER sockets as "clients" and REP and ROUTER sockets as "servers". Mostly, you'll want to bind REP and ROUTER sockets, and connect REQ and DEALER sockets to them. It's not always going to be this simple, but it is a clean and memorable place to start.

func (s *Server) Start() (err error) {
	s.Stateswitch <- STATE_RUNNING
	go func(s *Server) {
		s.Log.Println("Starting server...")
		s.Context, err = zmq.NewContext()
		if err != nil {

		}
		s.Socket, err = s.Context.NewSocket(zmq.ROUTER)
		if err != nil {

		}
		s.Socket.SetSockOptInt(zmq.LINGER, 0)
		s.Socket.Bind("tcp://127.0.0.1:6000")

		s.Log.Println("Server started...")

		s.PollItems = zmq.PollItems{
			zmq.PollItem{Socket: s.Socket, Events: zmq.POLLIN},
		}

		var cycleLength int64
		var cycleStart time.Time
		stateTicker := time.Tick(s.CheckStateChange)
	mainLoop:
		for {
			cycleStart = time.Now()
			// cycleLength = 0

			// make the poller run in a sep goroutine and push to a channel?

			if s.ActiveClients > 0 {
				_, _ = zmq.Poll(s.PollItems, 1)
			} else {
				// _, _ = zmq.Poll(s.PollItems, -1)
				_, _ = zmq.Poll(s.PollItems, 1000000)
			}
			switch {
			case s.PollItems[0].REvents&zmq.POLLIN != 0:
				s.Lock()
				msgParts, _ := s.PollItems[0].Socket.RecvMultipart(0)
				s.ActiveClients += 1
				s.Unlock()
				go s.HandleRequest(msgParts)
			default:
				select {
				case <-stateTicker:
					select {
					case s.State = <-s.Stateswitch:
						s.Log.Println("state changed:", s.State)
						if s.State == STATE_STOP {
							break mainLoop
						}
						if s.State == STATE_RUNNING {
						}
					default:
						time.Sleep(1)
					}
				default:
					// s.Log.Println(".")
				}
			}
			cycleLength = time.Now().UnixNano() - cycleStart.UnixNano()
			if cycleLength < s.CycleLength {
				// s.Log.Println("Sleep for ns", s.CycleLength-cycleLength)
				d := (s.CycleLength - cycleLength)
				time.Sleep(time.Duration(d) * time.Nanosecond)
			}
		}
		s.Log.Println("Server stopped running...")
		s.State = STATE_STOPPED
	}(s)
	s.Log.Println("Server starting...")
	return
}

func (s *Server) dbExists(name string) bool {
	if _, exists := s.Databases[name]; exists {
		return false
	}
	return true
}

func splitMsgs(payload []byte) (cmds [][]byte, args [][]byte, err error) {
	msgs := bytes.Split(bytes.Trim(payload, " "), []byte("\n"))
	for n, msg := range msgs {
		parts := bytes.Split(bytes.Trim(msg, " "), []byte(" "))
		// log.Println("parts:", parts)
		for i, p := range parts {
			if len(p) == 0 {
				continue
			}
			if i == 0 {
				cmds = append(cmds, p)
				args = append(args, []byte{})
			} else {
				args[n] = append(args[n], p...)
			}
		}
	}
	return
}

func (s *Server) HandleRequest(msgParts [][]byte) {
	c := NewClient(s, msgParts[0])
	cmds, args, err := splitMsgs(msgParts[2])
	if err != nil {
		// TODO
	}
	s.Log.Println("cmds, args:", cmds, args)

	// var retCode int
	var reply *Reply
	var replies []*Reply

	for i, cmd := range cmds {
		var cmdName string = strings.ToUpper(string(cmd))
		if _, exists := s.Commands[cmdName]; !exists {
			// handle non existing command call
			reply = NewReply(
				[][]byte{[]byte(fmt.Sprintf("Unknown Command %s.", string(cmd)))},
				COMMAND_FAIL)
			s.Log.Println(len(reply.Payload), reply)
		} else {
			reply = s.Commands[cmdName].Function(s, c, args[i])
			if reply.ReturnCode != COMMAND_OK {
				// s.Log.Println(string(reply.Payload))
			}
		}
		replies = append(replies, reply)
	}
	for rn, rep := range replies {
		c.Response = append(c.Response, rep.Serialize()...)
		if rn > 0 {
			c.Response = append(c.Response, []byte("\n")...)
		}
	}
	s.Lock()
	s.PollItems[0].Socket.SendMultipart([][]byte{c.Id, []byte(""), c.Response}, 0)

	// stats
	s.ActiveClients -= 1
	s.CommandsProcessed += 1
	s.Unlock()
	// TODO: count db write operations
}

/*
Take care of stuff...
*/
func (s *Server) prepareShutdown() {
}

func (s *Server) Stop() {
	s.Log.Println("Stopping server.")
	s.Stateswitch <- STATE_STOP
	for s.State != STATE_STOPPED {
		time.Sleep(100 * time.Millisecond)
	}
	s.Log.Println("Server teardown.")
	s.Socket.Close()
	s.Log.Println("Socket closed.")
	s.Context.Close()
	s.Log.Println("Context closed.")
	s.Log.Println("Stopped server.")
}
