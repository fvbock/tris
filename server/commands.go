package tris

import (
	"encoding/binary"
	"fmt"
	"github.com/fvbock/trie"
	"sort"
	"strings"
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
	ResponseLength() int64
	// ResponseSignature() []int
	Help() string
}

// make those "singletons"?

/*
CommandInfo sets the actuve database on the server client (the connection)
*/
type CommandInfo struct{}

func (cmd *CommandInfo) Name() string             { return "INFO" }
func (cmd *CommandInfo) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandInfo) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandInfo) ResponseLength() int64    { return 1 }
func (cmd *CommandInfo) ResponseSignature() []int { return []int{REPLY_TYPE_STRING} }
func (cmd *CommandInfo) Help() string             { return "TODO: CommandInfo text" }
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

	reply = NewReply([][]byte{[]byte(fmt.Sprintf("SERVER\n%v\nCLIENT\n%s", serverStr, c))}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	return
}

/*
CommandDbInfo sets the actuve database on the server client (the connection)
*/
type CommandDbInfo struct{}

func (cmd *CommandDbInfo) Name() string             { return "DBINFO" }
func (cmd *CommandDbInfo) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandDbInfo) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandDbInfo) ResponseLength() int64    { return 1 }
func (cmd *CommandDbInfo) ResponseSignature() []int { return []int{REPLY_TYPE_STRING} }
func (cmd *CommandDbInfo) Help() string             { return "TODO: CommandDbInfo text" }
func (cmd *CommandDbInfo) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	dbInfo := fmt.Sprintf(`DBINFO for database %s:
 OpsCount: %v
 LastPersistOpsCount: %v
 PersistOpsLimit: %v
 LastPersistTime: %v
 PersistInterval: %v
`, c.ActiveDb.Name, c.ActiveDb.OpsCount, c.ActiveDb.LastPersistOpsCount, c.ActiveDb.PersistOpsLimit, c.ActiveDb.LastPersistTime, c.ActiveDb.PersistInterval)

	reply = NewReply([][]byte{[]byte(dbInfo)}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	return
}

/*
CommandExit sets the actuve database on the server client (the connection)
*/
type CommandExit struct{}

func (cmd *CommandExit) Name() string             { return "EXIT" }
func (cmd *CommandExit) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandExit) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandExit) ResponseLength() int64    { return 0 }
func (cmd *CommandExit) ResponseSignature() []int { return []int{} }
func (cmd *CommandExit) Help() string             { return "TODO: CommandExit text" }
func (cmd *CommandExit) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	s.InactiveClientIds <- string(c.Id)
	reply = NewReply([][]byte{[]byte("")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	return
}

/*
CommandPing sets the actuve database on the server client (the connection)
*/
type CommandPing struct{}

func (cmd *CommandPing) Name() string             { return "PING" }
func (cmd *CommandPing) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandPing) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandPing) ResponseLength() int64    { return 1 }
func (cmd *CommandPing) ResponseSignature() []int { return []int{REPLY_TYPE_STRING} }
func (cmd *CommandPing) Help() string             { return "TODO: CommandPing text" }
func (cmd *CommandPing) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	reply = NewReply([][]byte{[]byte("0")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	return
}

/*
CommandSelect sets the actuve database on the server client (the connection)
*/
type CommandSelect struct{}

func (cmd *CommandSelect) Name() string             { return "SELECT" }
func (cmd *CommandSelect) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandSelect) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandSelect) ResponseLength() int64    { return 0 }
func (cmd *CommandSelect) ResponseSignature() []int { return []int{} }
func (cmd *CommandSelect) Help() string             { return "TODO: CommandSelect text" }
func (cmd *CommandSelect) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	// name := string(args[0].([]byte))
	name := args[0].(string)
	if !s.dbExists(name) {
		err := fmt.Sprintf("Databases %s does not exist.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}
	c.ActiveDb = s.Databases[name]
	// c.ActiveDbName = name
	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandCreateTrie maps to trie.Trie()
*/
type CommandCreateTrie struct{}

func (cmd *CommandCreateTrie) Name() string             { return "CREATE" }
func (cmd *CommandCreateTrie) Flags() int               { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandCreateTrie) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandCreateTrie) ResponseLength() int64    { return 0 }
func (cmd *CommandCreateTrie) ResponseSignature() []int { return []int{} }
func (cmd *CommandCreateTrie) Help() string             { return "TODO: CommandCreateTrie text" }
func (cmd *CommandCreateTrie) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	// name := string(args[0].([]byte))
	name := args[0].(string)
	s.Lock()
	defer s.Unlock()
	if s.dbExists(name) {
		err := fmt.Sprintf("Databases %s has already been registered.", name)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}
	s.NewDatabase(name)
	err := s.Databases[name].Persist(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, name))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the new db %s: %v", name, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}
	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandAdd maps to Trie.Add()
*/
type CommandAdd struct{}

func (cmd *CommandAdd) Name() string             { return "ADD" }
func (cmd *CommandAdd) Flags() int               { return COMMAND_FLAG_WRITE }
func (cmd *CommandAdd) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandAdd) ResponseLength() int64    { return 1 }
func (cmd *CommandAdd) ResponseSignature() []int { return []int{REPLY_TYPE_INT} }
func (cmd *CommandAdd) Help() string             { return "TODO: CommandAdd text" }
func (cmd *CommandAdd) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	b := c.ActiveDb.Db.Add(key)
	return NewReply([][]byte{[]byte(encodeIntReply(b.Count))}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandDel maps to Trie.Del()
*/
type CommandDel struct{}

func (cmd *CommandDel) Name() string             { return "DEL" }
func (cmd *CommandDel) Flags() int               { return COMMAND_FLAG_WRITE }
func (cmd *CommandDel) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandDel) ResponseLength() int64    { return 1 }
func (cmd *CommandDel) ResponseSignature() []int { return []int{REPLY_TYPE_BOOL} }
func (cmd *CommandDel) Help() string             { return "TODO: CommandDel text" }
func (cmd *CommandDel) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	if c.ActiveDb.Db.Delete(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandHas maps to Trie.Has()
*/
type CommandHas struct{}

func (cmd *CommandHas) Name() string             { return "HAS" }
func (cmd *CommandHas) Flags() int               { return COMMAND_FLAG_READ }
func (cmd *CommandHas) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHas) ResponseLength() int64    { return 1 }
func (cmd *CommandHas) ResponseSignature() []int { return []int{REPLY_TYPE_BOOL} }
func (cmd *CommandHas) Help() string             { return "TODO: CommandHas text" }
func (cmd *CommandHas) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	if c.ActiveDb.Db.Has(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandHasCount maps to Trie.HasCount()
*/
type CommandHasCount struct{}

func (cmd *CommandHasCount) Name() string             { return "HASCOUNT" }
func (cmd *CommandHasCount) Flags() int               { return COMMAND_FLAG_READ }
func (cmd *CommandHasCount) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasCount) ResponseLength() int64    { return 1 }
func (cmd *CommandHasCount) ResponseSignature() []int { return []int{REPLY_TYPE_INT} }
func (cmd *CommandHasCount) Help() string             { return "TODO: CommandHasCount text" }
func (cmd *CommandHasCount) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	_, count := c.ActiveDb.Db.HasCount(key)
	return NewReply([][]byte{[]byte(encodeIntReply(count))}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandHasPrefix maps to Trie.HasPrefix()
*/
type CommandHasPrefix struct{}

func (cmd *CommandHasPrefix) Name() string             { return "HASPREFIX" }
func (cmd *CommandHasPrefix) Flags() int               { return COMMAND_FLAG_READ }
func (cmd *CommandHasPrefix) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHasPrefix) ResponseLength() int64    { return 1 }
func (cmd *CommandHasPrefix) ResponseSignature() []int { return []int{REPLY_TYPE_BOOL} }
func (cmd *CommandHasPrefix) Help() string             { return "TODO: CommandHasPrefix text" }
func (cmd *CommandHasPrefix) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	if c.ActiveDb.Db.HasPrefix(key) {
		return NewReply([][]byte{[]byte("TRUE")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	}
	return NewReply([][]byte{[]byte("FALSE")}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandTree maps to Trie.Dump()
*/
type CommandTree struct{}

func (cmd *CommandTree) Name() string             { return "TREE" }
func (cmd *CommandTree) Flags() int               { return COMMAND_FLAG_READ }
func (cmd *CommandTree) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandTree) ResponseLength() int64    { return 1 }
func (cmd *CommandTree) ResponseSignature() []int { return []int{REPLY_TYPE_STRING} }
func (cmd *CommandTree) Help() string             { return "TODO: CommandTree text" }
func (cmd *CommandTree) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	return NewReply([][]byte{[]byte(c.ActiveDb.Db.Dump())}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandMembers maps to Trie.Members()
*/
type CommandMembers struct{}

func (cmd *CommandMembers) Name() string             { return "MEMBERS" }
func (cmd *CommandMembers) Flags() int               { return COMMAND_FLAG_READ }
func (cmd *CommandMembers) ResponseType() int        { return COMMAND_REPLY_MULTI }
func (cmd *CommandMembers) ResponseLength() int64    { return 2 }
func (cmd *CommandMembers) ResponseSignature() []int { return []int{REPLY_TYPE_STRING, REPLY_TYPE_INT} }
func (cmd *CommandMembers) Help() string             { return "TODO: CommandMembers text" }
func (cmd *CommandMembers) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	var mrep [][]byte
	for _, m := range c.ActiveDb.Db.Members() {
		mrep = append(mrep, []byte(m.Value), encodeIntReply(m.Count))
		// count := make([]byte, 4)
		// _ = binary.PutVarint(count, m.Count)
		// mrep = append(mrep, []byte(m.Value), count)
		// mrep = append(mrep, []byte(m.Value), []byte(string(m.Count)))
		// s.Log.Println(string(m.Count), []byte(string(m.Count)))
	}

	return NewReply(mrep, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandPrefixMembers maps to Trie.PrefixMembers()
*/
type CommandPrefixMembers struct{}

func (cmd *CommandPrefixMembers) Name() string          { return "PREFIXMEMBERS" }
func (cmd *CommandPrefixMembers) Flags() int            { return COMMAND_FLAG_READ }
func (cmd *CommandPrefixMembers) ResponseType() int     { return COMMAND_REPLY_MULTI }
func (cmd *CommandPrefixMembers) ResponseLength() int64 { return 2 }
func (cmd *CommandPrefixMembers) ResponseSignature() []int {
	return []int{REPLY_TYPE_STRING, REPLY_TYPE_INT}
}
func (cmd *CommandPrefixMembers) Help() string { return "TODO: CommandPrefixMembers text" }
func (cmd *CommandPrefixMembers) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	key := args[0].(string)
	var mrep [][]byte
	for _, m := range c.ActiveDb.Db.PrefixMembers(key) {
		count := make([]byte, 4)
		_ = binary.PutVarint(count, m.Count)
		mrep = append(mrep, []byte(m.Value), count)
		// mrep = append(mrep, []byte(m.Value), []byte(string(m.Count)))
	}

	return NewReply(mrep, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandTiming toggles the ShowExecTime flag on a client
*/
type CommandTiming struct{}

func (cmd *CommandTiming) Name() string             { return "TIMING" }
func (cmd *CommandTiming) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandTiming) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandTiming) ResponseLength() int64    { return 0 }
func (cmd *CommandTiming) ResponseSignature() []int { return []int{} }
func (cmd *CommandTiming) Help() string             { return "TODO: CommandTiming text" }
func (cmd *CommandTiming) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	if c.ShowExecTime {
		c.ShowExecTime = false
	} else {
		c.ShowExecTime = true
	}
	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandImportDb imports a a database from a file into a new database
*/
type CommandImportDb struct{}

func (cmd *CommandImportDb) Name() string             { return "IMPORT" }
func (cmd *CommandImportDb) Flags() int               { return COMMAND_FLAG_ADMIN | COMMAND_FLAG_WRITE }
func (cmd *CommandImportDb) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandImportDb) ResponseLength() int64    { return 0 }
func (cmd *CommandImportDb) ResponseSignature() []int { return []int{} }
func (cmd *CommandImportDb) Help() string             { return "TODO: CommandImportDb text" }
func (cmd *CommandImportDb) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	filename := args[0].(string)
	dbname := args[1].(string)
	s.Lock()
	if s.dbExists(dbname) {
		err := fmt.Sprintf("Databases %s already exists.", dbname)
		s.Log.Println(err)
		s.Unlock()
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}
	s.NewDatabase(dbname)
	s.Unlock()
	s.Databases[dbname].Db.Root.Lock()
	defer s.Databases[dbname].Db.Root.Unlock()
	var err error
	s.Databases[dbname].Db, err = trie.LoadFromFile(filename)
	if err != nil {
		err := fmt.Sprintf("Database import failed: %v", err)
		s.Log.Println(err)
		delete(s.Databases, dbname)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}

	// persist the db
	err = s.Databases[dbname].Persist(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, dbname))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the imported db %s: %v", dbname, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}
	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandMergeDb merges a a database from a file into a new database
*/
type CommandMergeDb struct{}

func (cmd *CommandMergeDb) Name() string             { return "MERGE" }
func (cmd *CommandMergeDb) Flags() int               { return COMMAND_FLAG_WRITE }
func (cmd *CommandMergeDb) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandMergeDb) ResponseLength() int64    { return 0 }
func (cmd *CommandMergeDb) ResponseSignature() []int { return []int{} }
func (cmd *CommandMergeDb) Help() string             { return "TODO: CommandMergeDb text" }
func (cmd *CommandMergeDb) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	filename := args[0].(string)
	err := c.ActiveDb.Db.MergeFromFile(filename)
	if err != nil {
		err := fmt.Sprintf("Database merge failed: %v", err)
		s.Log.Println(err)
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}

	// backup?

	// persist the db
	err = s.Databases[c.ActiveDb.Name].Persist(fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, c.ActiveDb.Name))
	if err != nil {
		errMsg := fmt.Sprintf("Could persist the mergeed db %s: %v", c.ActiveDb.Name, err)
		s.Log.Println(errMsg)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}
	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandSave saves a full Trie to disk in a separate process
*/
type CommandSave struct{}

func (cmd *CommandSave) Name() string             { return "SAVE" }
func (cmd *CommandSave) Flags() int               { return COMMAND_FLAG_WRITE }
func (cmd *CommandSave) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandSave) ResponseLength() int64    { return 0 }
func (cmd *CommandSave) ResponseSignature() []int { return []int{} }
func (cmd *CommandSave) Help() string             { return "TODO: CommandSave text" }
func (cmd *CommandSave) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	if c.ActiveDb.Name == DEFAULT_DB {
		err := fmt.Sprintf("Manually saving the default DB is not permitted.")
		return NewReply([][]byte{[]byte(err)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}

	srcFilePath := fmt.Sprintf("%s/%s%s", s.Config.DataDir, s.Config.StorageFilePrefix, c.ActiveDb.Name)
	dstPath := fmt.Sprintf("%s/%s_bak", s.Config.DataDir, c.ActiveDb.Name)
	err := c.ActiveDb.Backup(srcFilePath, dstPath, c.ActiveDb.Name)
	if err != nil {
		errMsg := fmt.Sprintf("Backup failed: %v", err)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}

	err = c.ActiveDb.Persist(srcFilePath)
	if err != nil {
		errMsg := fmt.Sprintf("Backup failed: %v", err)
		return NewReply([][]byte{[]byte(errMsg)}, COMMAND_FAIL, 1, cmd.ResponseSignature())
	}

	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandShutdown shuts down the server
*/
type CommandShutdown struct{}

func (cmd *CommandShutdown) Name() string             { return "SHUTDOWN" }
func (cmd *CommandShutdown) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandShutdown) ResponseType() int        { return COMMAND_REPLY_EMPTY }
func (cmd *CommandShutdown) ResponseLength() int64    { return 0 }
func (cmd *CommandShutdown) ResponseSignature() []int { return []int{} }
func (cmd *CommandShutdown) Help() string             { return "TODO: CommandShutdown text" }
func (cmd *CommandShutdown) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	s.Stop()
	return NewReply([][]byte{}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
}

/*
CommandHelp shuts down the server
*/
type CommandHelp struct{}

func (cmd *CommandHelp) Name() string             { return "HELP" }
func (cmd *CommandHelp) Flags() int               { return COMMAND_FLAG_ADMIN }
func (cmd *CommandHelp) ResponseType() int        { return COMMAND_REPLY_SINGLE }
func (cmd *CommandHelp) ResponseLength() int64    { return 1 }
func (cmd *CommandHelp) ResponseSignature() []int { return []int{REPLY_TYPE_STRING} }
func (cmd *CommandHelp) Help() string             { return "HELP help: TODO" }
func (cmd *CommandHelp) Function(s *Server, c *ClientConnection, args ...interface{}) (reply *Reply) {
	cmdName := strings.ToUpper(args[0].(string))
	if _, exists := s.Commands[cmdName]; !exists {
		reply = NewReply([][]byte{[]byte(fmt.Sprintf("Unknown Command %s.\n\n%s", cmdName, cmd.Help()))}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	} else {
		reply = NewReply([][]byte{[]byte(fmt.Sprintf("\n%s\n", s.Commands[cmdName].Help()))}, COMMAND_OK, cmd.ResponseLength(), cmd.ResponseSignature())
	}
	return
}
