package tris

import (
	"fmt"
	"github.com/fvbock/trie"
	"github.com/fvbock/tris/util"
	"os"
	"strconv"
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

func (cmd *CommandInfo) Name() string      { return "INFO" }
func (cmd *CommandInfo) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandInfo) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandInfo) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	var dbNames string
	var n int = 1
	for name, _ := range s.Databases {
		if name != DEFAULT_DB {
			dbNames += fmt.Sprintf("    %v) %s\n", n, name)
			n += 1
		}
	}
	serverStr := fmt.Sprintf(`Tris %s.
Host: %s
Port: %v
DataDir: %s

Databases:
  Default DB: %s
  User DBs:
%v

ActiveClients: %v
Commands Processed: %v
Commands Running: %v
`, VERSION, s.Config.Host, s.Config.Port, s.Config.DataDir, DEFAULT_DB, dbNames, len(s.ActiveClients), s.CommandsProcessed, s.RequestsRunning)

	reply = NewReply([][]byte{[]byte(fmt.Sprintf("SERVER\n%v\nCLIENT\n%s", serverStr, c))}, COMMAND_OK)
	return
}

/*
CommandDbInfo sets the actuve database on the server client (the connection)
*/
type CommandDbInfo struct{}

func (cmd *CommandDbInfo) Name() string      { return "DBINFO" }
func (cmd *CommandDbInfo) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandDbInfo) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandDbInfo) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	var dbNames string
	var n int = 1
	for name, _ := range s.Databases {
		if name != DEFAULT_DB {
			dbNames += fmt.Sprintf("    %v) %s\n", n, name)
			n += 1
		}
	}
	serverStr := fmt.Sprintf(`Tris %s.
Host: %s
Port: %v
DataDir: %s

Databases:
  Default DB: %s
  User DBs:
%v

ActiveClients: %v
Commands Processed: %v
Commands Running: %v
`, VERSION, s.Config.Host, s.Config.Port, s.Config.DataDir, DEFAULT_DB, dbNames, len(s.ActiveClients), s.CommandsProcessed, s.RequestsRunning)

	reply = NewReply([][]byte{[]byte(fmt.Sprintf("SERVER\n%v\nCLIENT\n%s", serverStr, c))}, COMMAND_OK)
	return
}

/*
CommandExit sets the actuve database on the server client (the connection)
*/
type CommandExit struct{}

func (cmd *CommandExit) Name() string      { return "EXIT" }
func (cmd *CommandExit) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandExit) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandExit) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// delete(s.ActiveClients, string(c.Id))
	s.InactiveClientIds <- string(c.Id)
	reply = NewReply([][]byte{[]byte("")}, COMMAND_OK)
	return
}

/*
CommandPing sets the actuve database on the server client (the connection)
*/
type CommandPing struct{}

func (cmd *CommandPing) Name() string      { return "PING" }
func (cmd *CommandPing) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandPing) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandPing) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	reply = NewReply([][]byte{[]byte("0")}, COMMAND_OK)
	return
}

/*
CommandSelect sets the actuve database on the server client (the connection)
*/
type CommandSelect struct{}

func (cmd *CommandSelect) Name() string      { return "SELECT" }
func (cmd *CommandSelect) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandSelect) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandSelect) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// name := string(args[0].([]byte))
	name := args[0].(string)
	if !s.dbExists(name) {
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

func (cmd *CommandCreateTrie) Name() string      { return "CREATE" }
func (cmd *CommandCreateTrie) Flags() int        { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandCreateTrie) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandCreateTrie) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// name := string(args[0].([]byte))
	name := args[0].(string)
	s.Lock()
	defer s.Unlock()
	if s.dbExists(name) {
		err := fmt.Sprintf("Databases %s has already been registered.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	s.Databases[name] = trie.NewRefCountTrie()
	err := s.Databases[name].DumpToFile(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, name))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the new db %s: %v", name, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandAdd maps to RefCountTrie.Add()
*/
type CommandAdd struct{}

func (cmd *CommandAdd) Name() string      { return "ADD" }
func (cmd *CommandAdd) Flags() int        { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandAdd) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandAdd) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// key := string(args[0].([]byte))
	key := args[0].(string)
	b := c.ActiveDb.Add(key)
	return NewReply([][]byte{[]byte(string(b.Count))}, COMMAND_OK)
}

/*
CommandDel maps to RefCountTrie.Del()
*/
type CommandDel struct{}

func (cmd *CommandDel) Name() string      { return "DEL" }
func (cmd *CommandDel) Flags() int        { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandDel) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandDel) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// key := string(args[0].([]byte))
	key := args[0].(string)
	if c.ActiveDb.Delete(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandHas maps to RefCountTrie.Has()
*/
type CommandHas struct{}

func (cmd *CommandHas) Name() string      { return "HAS" }
func (cmd *CommandHas) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandHas) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHas) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// key := string(args[0].([]byte))
	key := args[0].(string)
	if c.ActiveDb.Has(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandHasCount maps to RefCountTrie.HasCount()
*/
type CommandHasCount struct{}

func (cmd *CommandHasCount) Name() string      { return "HASCOUNT" }
func (cmd *CommandHasCount) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandHasCount) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasCount) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// key := string(args[0].([]byte))
	key := args[0].(string)
	has, count := c.ActiveDb.HasCount(key)
	s.Log.Println(has, count, string(count))
	return NewReply([][]byte{[]byte(strconv.FormatInt(int64(count), 10))}, COMMAND_OK)
}

/*
CommandHasPrefix maps to RefCountTrie.HasPrefix()
*/
type CommandHasPrefix struct{}

func (cmd *CommandHasPrefix) Name() string      { return "HASPREFIX" }
func (cmd *CommandHasPrefix) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandHasPrefix) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasPrefix) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// key := string(args[0].([]byte))
	key := args[0].(string)
	if c.ActiveDb.HasPrefix(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandTree maps to RefCountTrie.Dump()
*/
type CommandTree struct{}

func (cmd *CommandTree) Name() string      { return "TREE" }
func (cmd *CommandTree) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandTree) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandTree) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	return NewReply([][]byte{[]byte(c.ActiveDb.Dump())}, COMMAND_OK)
}

/*
CommandMembers maps to RefCountTrie.Members()
*/
type CommandMembers struct{}

func (cmd *CommandMembers) Name() string      { return "MEMBERS" }
func (cmd *CommandMembers) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandMembers) ResponseType() int { return COMMAND_REPLY_MULTI }
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

func (cmd *CommandPrefixMembers) Name() string      { return "PREFIXMEMBERS" }
func (cmd *CommandPrefixMembers) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandPrefixMembers) ResponseType() int { return COMMAND_REPLY_MULTI }
func (cmd *CommandPrefixMembers) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	// key := string(args[0].([]byte))
	key := args[0].(string)
	var mrep [][]byte
	for _, m := range c.ActiveDb.PrefixMembers(key) {
		mrep = append(mrep, []byte(m.Value))
	}

	return NewReply(mrep, COMMAND_OK)
}

/*
CommandTiming toggles the ShowExecTime flag on a client
*/
type CommandTiming struct{}

func (cmd *CommandTiming) Name() string      { return "TIMING" }
func (cmd *CommandTiming) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandTiming) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandTiming) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	if c.ShowExecTime {
		c.ShowExecTime = false
	} else {
		c.ShowExecTime = true
	}
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandImportDb imports a a database from a file into a new database
*/
type CommandImportDb struct{}

func (cmd *CommandImportDb) Name() string      { return "IMPORT" }
func (cmd *CommandImportDb) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandImportDb) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandImportDb) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	filename := args[0].(string)
	dbname := args[1].(string)
	s.Lock()
	if s.dbExists(dbname) {
		err := fmt.Sprintf("Databases %s already exists.", dbname)
		s.Log.Println(err)
		s.Unlock()
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	s.Databases[dbname] = trie.NewRefCountTrie()
	s.Unlock()
	s.Databases[dbname].Root.Lock()
	defer s.Databases[dbname].Root.Unlock()
	tr, err := trie.RCTLoadFromFile(filename)
	if err != nil {
		err := fmt.Sprintf("Database import failed: %v", err)
		s.Log.Println(err)
		delete(s.Databases, dbname)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	s.Databases[dbname] = tr

	// persist the db
	err = s.Databases[dbname].DumpToFile(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, dbname))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the imported db %s: %v", dbname, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandSave saves a full Trie to disk in a separate process
*/
type CommandSave struct{}

func (cmd *CommandSave) Name() string      { return "SAVE" }
func (cmd *CommandSave) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandSave) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandSave) Function(s *Server, c *Client, args ...interface{}) (reply *Reply) {
	if c.ActiveDbName == DEFAULT_DB {
		err := fmt.Sprintf("Manually saving the default DB is not permitted.")
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	dbBackupPath := fmt.Sprintf("%s/%s_bak", s.Config.DataDir, c.ActiveDbName)
	exists, err := tris.PathExists(dbBackupPath)
	if !exists {
		if err != nil {
			errMsg := fmt.Sprintf("Could not stat directory %s for backup files: %v", dbBackupPath, err)
			s.Log.Println(errMsg)
			return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
		} else {
			err = os.Mkdir(dbBackupPath, 0777)
			if err != nil {
				errMsg := fmt.Sprintf("Could not create directory %s for backup files: %v", dbBackupPath, err)
				s.Log.Println(errMsg)
				return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
			}
		}
	}

	// copy the old file into the backup folder
	// TODO: this should be dropped and be replaced by a write ops log + timestamp
	srcFile := fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, c.ActiveDbName)
	dstFile := fmt.Sprintf("%s/%s", dbBackupPath, c.ActiveDbName)
	c.ActiveDb.Root.Lock()
	err = tris.CopyFile(srcFile, dstFile)
	c.ActiveDb.Root.Unlock()
	if err != nil {
		errMsg := fmt.Sprintf("Could backup the previous data file: %v", err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}

	err = c.ActiveDb.DumpToFile(srcFile)
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the db %s: %v", c.ActiveDbName, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}

	return NewReply([][]byte{}, COMMAND_OK)
}
