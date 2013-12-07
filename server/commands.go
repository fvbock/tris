package tris

import (
	"fmt"
	"github.com/fvbock/trie"
	"sort"
	"strconv"
)

const (
	COMMAND_FLAG_READ  = 1
	COMMAND_FLAG_WRITE = 2
	COMMAND_FLAG_ADMIN = 4

	COMMAND_REPLY_EMPTY  = 0
	COMMAND_REPLY_SINGLE = 1
	COMMAND_REPLY_MULTI  = 2
	// COMMAND_REPLY_NONE   = 3

	COMMAND_OK   = 0
	COMMAND_FAIL = 1
)

var (
	TrisCommands []Command
)

/*
Server command interface:

Name() will return the name by which the command is identified in the servers command table
Function() will be the actual function executed by calling the command. all functions get the Server and executing ClientConnection passed as pointers.
*/
type Command interface {
	Name() string
	Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply)
	Flags() int
	ResponseType() int
	Help() string
}

// make those "singletons"?

/*
CommandInfo sets the actuve database on the server client (the connection)
*/
type CommandInfo struct{}

func (cmd *CommandInfo) Name() string      { return "INFO" }
func (cmd *CommandInfo) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandInfo) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandInfo) Help() string      { return "TODO" }
func (cmd *CommandInfo) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	var dbNames sort.StringSlice
	var dbList string
	for name, _ := range s.Databases {
		if name != DEFAULT_DB {
			dbNames = append(dbNames, name)
		}
	}
	sort.Sort(dbNames)
	var n int = 1
	var marker string
	for _, name := range dbNames {
		if name == c.ActiveDb.Name {
			marker = "* "
		} else {
			marker = ""
		}
		dbList += fmt.Sprintf("    %v) %s%s\n", n, marker, name)
		n += 1
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
`, VERSION, s.Config.Host, s.Config.Port, s.Config.DataDir, DEFAULT_DB, dbList, len(s.ActiveClients), s.CommandsProcessed, s.RequestsRunning)

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
func (cmd *CommandDbInfo) Help() string      { return "TODO" }
func (cmd *CommandDbInfo) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	dbInfo := fmt.Sprintf(`DBINFO for database %s:
 OpsCount: %v
 LastPersistOpsCount: %v
 PersistOpsLimit: %v
 LastPersistTime: %v
 PersistInterval: %v
`, c.ActiveDb.Name, c.ActiveDb.OpsCount, c.ActiveDb.LastPersistOpsCount, c.ActiveDb.PersistOpsLimit, c.ActiveDb.LastPersistTime, c.ActiveDb.PersistInterval)

	reply = NewReply([][]byte{[]byte(dbInfo)}, COMMAND_OK)
	return
}

/*
CommandExit sets the actuve database on the server client (the connection)
*/
type CommandExit struct{}

func (cmd *CommandExit) Name() string      { return "EXIT" }
func (cmd *CommandExit) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandExit) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandExit) Help() string      { return "TODO" }
func (cmd *CommandExit) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
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
func (cmd *CommandPing) Help() string      { return "TODO" }
func (cmd *CommandPing) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
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
func (cmd *CommandSelect) Help() string      { return "TODO" }
func (cmd *CommandSelect) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	// name := string(args[0].([]byte))
	name := args[0].(string)
	if !s.dbExists(name) {
		err := fmt.Sprintf("Databases %s does not exist.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	c.ActiveDb = s.Databases[name]
	// c.ActiveDbName = name
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandCreateTrie maps to trie.Trie()
*/
type CommandCreateTrie struct{}

func (cmd *CommandCreateTrie) Name() string      { return "CREATE" }
func (cmd *CommandCreateTrie) Flags() int        { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandCreateTrie) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandCreateTrie) Help() string      { return "TODO" }
func (cmd *CommandCreateTrie) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	// name := string(args[0].([]byte))
	name := args[0].(string)
	s.Lock()
	defer s.Unlock()
	if s.dbExists(name) {
		err := fmt.Sprintf("Databases %s has already been registered.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	s.NewDatabase(name)
	err := s.Databases[name].Persist(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, name))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the new db %s: %v", name, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandAdd maps to Trie.Add()
*/
type CommandAdd struct{}

func (cmd *CommandAdd) Name() string      { return "ADD" }
func (cmd *CommandAdd) Flags() int        { return COMMAND_FLAG_WRITE }
func (cmd *CommandAdd) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandAdd) Help() string      { return "TODO" }
func (cmd *CommandAdd) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	b := c.ActiveDb.Db.Add(key)
	return NewReply([][]byte{[]byte(string(b.Count))}, COMMAND_OK)
}

/*
CommandDel maps to Trie.Del()
*/
type CommandDel struct{}

func (cmd *CommandDel) Name() string      { return "DEL" }
func (cmd *CommandDel) Flags() int        { return COMMAND_FLAG_WRITE }
func (cmd *CommandDel) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandDel) Help() string      { return "TODO" }
func (cmd *CommandDel) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	if c.ActiveDb.Db.Delete(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandHas maps to Trie.Has()
*/
type CommandHas struct{}

func (cmd *CommandHas) Name() string      { return "HAS" }
func (cmd *CommandHas) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandHas) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHas) Help() string      { return "TODO" }
func (cmd *CommandHas) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	if c.ActiveDb.Db.Has(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandHasCount maps to Trie.HasCount()
*/
type CommandHasCount struct{}

func (cmd *CommandHasCount) Name() string      { return "HASCOUNT" }
func (cmd *CommandHasCount) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandHasCount) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasCount) Help() string      { return "TODO" }
func (cmd *CommandHasCount) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	_, count := c.ActiveDb.Db.HasCount(key)
	return NewReply([][]byte{[]byte(strconv.FormatInt(int64(count), 10))}, COMMAND_OK)
}

/*
CommandHasPrefix maps to Trie.HasPrefix()
*/
type CommandHasPrefix struct{}

func (cmd *CommandHasPrefix) Name() string      { return "HASPREFIX" }
func (cmd *CommandHasPrefix) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandHasPrefix) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasPrefix) Help() string      { return "TODO" }
func (cmd *CommandHasPrefix) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	if c.ActiveDb.Db.HasPrefix(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK)
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK)
}

/*
CommandTree maps to Trie.Dump()
*/
type CommandTree struct{}

func (cmd *CommandTree) Name() string      { return "TREE" }
func (cmd *CommandTree) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandTree) ResponseType() int { return COMMAND_REPLY_SINGLE }
func (cmd *CommandTree) Help() string      { return "TODO" }
func (cmd *CommandTree) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	return NewReply([][]byte{[]byte(c.ActiveDb.Db.Dump())}, COMMAND_OK)
}

/*
CommandMembers maps to Trie.Members()
*/
type CommandMembers struct{}

func (cmd *CommandMembers) Name() string      { return "MEMBERS" }
func (cmd *CommandMembers) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandMembers) ResponseType() int { return COMMAND_REPLY_MULTI }
func (cmd *CommandMembers) Help() string      { return "TODO" }
func (cmd *CommandMembers) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	var mrep [][]byte
	for _, m := range c.ActiveDb.Db.Members() {
		mrep = append(mrep, []byte(m.Value))
	}

	return NewReply(mrep, COMMAND_OK)
}

/*
CommandPrefixMembers maps to Trie.PrefixMembers()
*/
type CommandPrefixMembers struct{}

func (cmd *CommandPrefixMembers) Name() string      { return "PREFIXMEMBERS" }
func (cmd *CommandPrefixMembers) Flags() int        { return COMMAND_FLAG_READ }
func (cmd *CommandPrefixMembers) ResponseType() int { return COMMAND_REPLY_MULTI }
func (cmd *CommandPrefixMembers) Help() string      { return "TODO" }
func (cmd *CommandPrefixMembers) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	var mrep [][]byte
	for _, m := range c.ActiveDb.Db.PrefixMembers(key) {
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
func (cmd *CommandTiming) Help() string      { return "TODO" }
func (cmd *CommandTiming) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
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
func (cmd *CommandImportDb) Flags() int        { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandImportDb) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandImportDb) Help() string      { return "TODO" }
func (cmd *CommandImportDb) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	filename := args[0].(string)
	dbname := args[1].(string)
	s.Lock()
	if s.dbExists(dbname) {
		err := fmt.Sprintf("Databases %s already exists.", dbname)
		s.Log.Println(err)
		s.Unlock()
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}
	s.NewDatabase(dbname)
	s.Unlock()
	s.Databases[dbname].Db.Root.Lock()
	defer s.Databases[dbname].Db.Root.Unlock()
	var err error
	s.Databases[dbname].Db, err = trie.RCTLoadFromFile(filename)
	if err != nil {
		err := fmt.Sprintf("Database import failed: %v", err)
		s.Log.Println(err)
		delete(s.Databases, dbname)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}

	// persist the db
	err = s.Databases[dbname].Persist(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, dbname))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the imported db %s: %v", dbname, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}
	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandMergeDb merges a a database from a file into a new database
*/
type CommandMergeDb struct{}

func (cmd *CommandMergeDb) Name() string      { return "MERGE" }
func (cmd *CommandMergeDb) Flags() int        { return COMMAND_FLAG_WRITE }
func (cmd *CommandMergeDb) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandMergeDb) Help() string      { return "TODO" }
func (cmd *CommandMergeDb) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	filename := args[0].(string)
	err := c.ActiveDb.Db.RCTMergeFromFile(filename)
	if err != nil {
		err := fmt.Sprintf("Database merge failed: %v", err)
		s.Log.Println(err)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}

	// backup?

	// persist the db
	err = s.Databases[c.ActiveDb.Name].Persist(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, c.ActiveDb.Name))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the mergeed db %s: %v", c.ActiveDb.Name, err)
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
func (cmd *CommandSave) Flags() int        { return COMMAND_FLAG_WRITE }
func (cmd *CommandSave) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandSave) Help() string      { return "TODO" }
func (cmd *CommandSave) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	if c.ActiveDb.Name == DEFAULT_DB {
		err := fmt.Sprintf("Manually saving the default DB is not permitted.")
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL)
	}

	srcFilePath := fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, c.ActiveDb.Name)
	dstPath := fmt.Sprintf("%s/%s_bak", s.Config.DataDir, c.ActiveDb.Name)
	err := c.ActiveDb.Backup(srcFilePath, dstPath, c.ActiveDb.Name)
	if err != nil {
		errMsg := fmt.Sprintf("Backup failed: %v", err)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}

	err = c.ActiveDb.Persist(srcFilePath)
	if err != nil {
		errMsg := fmt.Sprintf("Backup failed: %v", err)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL)
	}

	return NewReply([][]byte{}, COMMAND_OK)
}

/*
CommandShutdown shuts down the server
*/
type CommandShutdown struct{}

func (cmd *CommandShutdown) Name() string      { return "SHUTDOWN" }
func (cmd *CommandShutdown) Flags() int        { return COMMAND_FLAG_ADMIN }
func (cmd *CommandShutdown) ResponseType() int { return COMMAND_REPLY_EMPTY }
func (cmd *CommandShutdown) Help() string      { return "TODO" }
func (cmd *CommandShutdown) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	s.Stop()
	return NewReply([][]byte{}, COMMAND_OK)
}
