package tris

import (
	"fmt"
	// "github.com/fvbock/trie"
)

type ClientConnection struct {
	Id           []byte
	Msg          []byte
	ActiveDb     *Database
	ShowExecTime bool
}

func (c *ClientConnection) String() string {
	return fmt.Sprintf("Client ID: %v\nActive Db: %v\n", c.Id, c.ActiveDb.Name)
}

func NewClientConnection(s *Server, id []byte) *ClientConnection {
	return &ClientConnection{
		Id:           id,
		ActiveDb:     s.Databases[DEFAULT_DB],
		ShowExecTime: false,
	}
}
