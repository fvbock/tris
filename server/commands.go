package tris

import (
	// "errors"
	"fmt"
	"github.com/fvbock/trie"
	"time"
)

var (
	TrisCommands []Command
)

func init() {
	TrisCommands = append(TrisCommands, &CommandInfo{})
	TrisCommands = append(TrisCommands, &CommandSelect{})
	TrisCommands = append(TrisCommands, &CommandCreateTrie{})
	TrisCommands = append(TrisCommands, &CommandAdd{})
}

// make those "singletons"?

/*
CommandInfo sets the actuve database on the server client (the connection)
*/
type CommandInfo struct{}

func (cmd *CommandInfo) Name() string { return "INFO" }
func (cmd *CommandInfo) Flags() int   { return COMMAND_FLAG_ADMIN }
func (cmd *CommandInfo) Function(s *Server, c *Client, args ...interface{}) (retCode int, reply *Reply) {
	reply = NewReply([]byte(fmt.Sprintf("%v Tris 0.0.1 says Hello and INFO.\nServer: %v\nClient: %v\n", time.Now(), s, c)), COMMAND_OK)
	return COMMAND_OK, reply
}

/*
CommandSelect sets the actuve database on the server client (the connection)
*/
type CommandSelect struct{}

func (cmd *CommandSelect) Name() string { return "SELECT" }
func (cmd *CommandSelect) Flags() int   { return COMMAND_FLAG_ADMIN }
func (cmd *CommandSelect) Function(s *Server, c *Client, args ...interface{}) (retCode int, reply *Reply) {
	name := string(args[0].([]byte))
	if _, exists := s.Databases[name]; exists {
		err := fmt.Sprintf("Databases %s does not exist.", name)
		return COMMAND_FAIL, NewReply([]byte(err), COMMAND_FAIL)
	}
	c.ActiveDb = s.Databases[name]
	c.ActiveDbName = name
	return COMMAND_OK, NewReply([]byte{}, COMMAND_OK)
}

/*
CommandCreateTrie maps to trie.NewRefCountTrie()
*/
type CommandCreateTrie struct{}

func (cmd *CommandCreateTrie) Name() string { return "CREATE" }
func (cmd *CommandCreateTrie) Flags() int   { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandCreateTrie) Function(s *Server, c *Client, args ...interface{}) (retCode int, reply *Reply) {
	name := args[0].(string)
	if _, exists := s.Databases[name]; exists {
		err := fmt.Sprintf("Databases %s has already been registered.", name)
		return COMMAND_FAIL, NewReply([]byte(err), COMMAND_FAIL)
	}
	s.Databases[name] = trie.NewRefCountTrie()
	return COMMAND_OK, NewReply([]byte{}, COMMAND_OK)
}

/*
CommandAdd maps to RefCountTrie.Add()
*/
type CommandAdd struct{}

func (cmd *CommandAdd) Name() string { return "ADD" }
func (cmd *CommandAdd) Flags() int   { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandAdd) Function(s *Server, c *Client, args ...interface{}) (retCode int, reply *Reply) {
	name := args[0].(string)
	if _, exists := s.Databases[name]; exists {
		err := fmt.Sprintf("Databases %s does not exist.", name)
		return COMMAND_FAIL, NewReply([]byte(err), COMMAND_FAIL)
	}
	s.Databases[name] = trie.NewRefCountTrie()
	return COMMAND_OK, NewReply([]byte{}, COMMAND_OK)
}
