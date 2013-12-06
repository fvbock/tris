package tris

import (
	"github.com/fvbock/trie"
)

type Database struct {
	Name                     string
	Db                       *trie.RefCountTrie
	OpsCount                 int
	DumpOpsCount             int
	PersistThresholdOpsCount int
	// PersistThresholdTime time.Duration
}

// TODO: move persisting into the Database struct
