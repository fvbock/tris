package tris

import (
	// "errors"
	"fmt"
	"github.com/fvbock/trie"
	"strconv"
	"time"
)

var (
	TrisCommands []Command
)

func init() {
}

// make those "singletons"?

/*
CommandInfo sets the actuve database on the server client (the connection)
*/
type CommandInfo struct{}

func (cmd *CommandInfo) Name() string       { return "INFO" }
func (cmd *CommandInfo) Flags() int         { return COMMAND_FLAG_ADMIN }
func (cmd *CommandInfo) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandInfo) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	reply = NewReply([][]byte{[]byte(fmt.Sprintf("%v Tris 0.0.1 says Hello and INFO.\nServer: %v\nClient: %v\n", time.Now(), s, c))}, COMMAND_OK)
	return
}

/*
CommandSelect sets the actuve database on the server client (the connection)
*/
type CommandSelect struct{}

func (cmd *CommandSelect) Name() string       { return "SELECT" }
func (cmd *CommandSelect) Flags() int         { return COMMAND_FLAG_ADMIN }
func (cmd *CommandSelect) ResponseFlags() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandSelect) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	name := string(args[0].([]byte))
	if _, exists := s.Databases[name]; exists {
		err := fmt.Sprintf("Databases %s does not exist.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	c.ActiveDb = s.Databases[name]
	c.ActiveDbName = name
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandCreateTrie maps to trie.NewRefCountTrie()
*/
type CommandCreateTrie struct{}

func (cmd *CommandCreateTrie) Name() string       { return "CREATE" }
func (cmd *CommandCreateTrie) Flags() int         { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandCreateTrie) ResponseFlags() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandCreateTrie) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	name := string(args[0].(uint8))
	if _, exists := s.Databases[name]; exists {
		err := fmt.Sprintf("Databases %s has already been registered.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	s.Databases[name] = trie.NewRefCountTrie()
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandAdd maps to RefCountTrie.Add()
*/
type CommandAdd struct{}

func (cmd *CommandAdd) Name() string       { return "ADD" }
func (cmd *CommandAdd) Flags() int         { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandAdd) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandAdd) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	key := string(args[0].([]uint8))
	b := c.ActiveDb.Add(key)
	return NewReply([][]byte{[]byte(string(b.Count))}, COMMAND_OK)
}

/*
CommandDel maps to RefCountTrie.Del()
*/
type CommandDel struct{}

func (cmd *CommandDel) Name() string       { return "DEL" }
func (cmd *CommandDel) Flags() int         { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandDel) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandDel) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	key := string(args[0].([]uint8))
	if c.ActiveDb.Delete(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandHas maps to RefCountTrie.Has()
*/
type CommandHas struct{}

func (cmd *CommandHas) Name() string       { return "HAS" }
func (cmd *CommandHas) Flags() int         { return COMMAND_FLAG_READ }
func (cmd *CommandHas) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHas) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	key := string(args[0].([]uint8))
	if c.ActiveDb.Has(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandHasCount maps to RefCountTrie.HasCount()
*/
type CommandHasCount struct{}

func (cmd *CommandHasCount) Name() string       { return "HASCOUNT" }
func (cmd *CommandHasCount) Flags() int         { return COMMAND_FLAG_READ }
func (cmd *CommandHasCount) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasCount) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	key := string(args[0].([]uint8))
	has, count := c.ActiveDb.HasCount(key)
	s.Log.Println(has, count, string(count))
	// return NewReply([][]byte{[]byte(string(count))}, COMMAND_OK)
	return NewReply([][]byte{[]byte(strconv.FormatInt(count, 10))}, COMMAND_OK)
}

/*
CommandHasPrefix maps to RefCountTrie.HasPrefix()
*/
type CommandHasPrefix struct{}

func (cmd *CommandHasPrefix) Name() string       { return "HASPREFIX" }
func (cmd *CommandHasPrefix) Flags() int         { return COMMAND_FLAG_READ }
func (cmd *CommandHasPrefix) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasPrefix) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	key := string(args[0].([]uint8))
	if c.ActiveDb.HasPrefix(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandTree maps to RefCountTrie.Dump()
*/
type CommandTree struct{}

func (cmd *CommandTree) Name() string       { return "TREE" }
func (cmd *CommandTree) Flags() int         { return COMMAND_FLAG_READ }
func (cmd *CommandTree) ResponseFlags() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandTree) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	return NewReply([][]byte{[]byte(c.ActiveDb.Dump())}, COMMAND_OK)
}

/*
CommandMembers maps to RefCountTrie.Members()
*/
type CommandMembers struct{}

func (cmd *CommandMembers) Name() string       { return "MEMBERS" }
func (cmd *CommandMembers) Flags() int         { return COMMAND_FLAG_READ }
func (cmd *CommandMembers) ResponseFlags() int { return COMMAND_REPLY_MULTI }
func (cmd *CommandMembers) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	var mrep [][]byte
	for _, m := range c.ActiveDb.Members() {
		mrep = append(mrep, []byte(m.Value))
	}

	return NewReply(mrep, COMMAND_OK)
}

/*
CommandPrefixMembers maps to RefCountTrie.PrefixMembers()
*/
type CommandPrefixMembers struct{}

func (cmd *CommandPrefixMembers) Name() string       { return "PREFIXMEMBERS" }
func (cmd *CommandPrefixMembers) Flags() int         { return COMMAND_FLAG_READ }
func (cmd *CommandPrefixMembers) ResponseFlags() int { return COMMAND_REPLY_MULTI }
func (cmd *CommandPrefixMembers) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	key := string(args[0].([]uint8))
	var mrep [][]byte
	for _, m := range c.ActiveDb.PrefixMembers(key) {
		mrep = append(mrep, []byte(m.Value))
	}

	return NewReply(mrep, COMMAND_OK)
}
