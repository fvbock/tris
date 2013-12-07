package tris

import (
	"errors"
	"fmt"
	"github.com/fvbock/trie"
	"github.com/fvbock/tris/util"
	"os"
	"sync"
	"time"
)

type Database struct {
	sync.RWMutex
	Name                string
	Db                  *trie.Trie
	OpsCount            int
	LastPersistOpsCount int
	PersistOpsLimit     int
	LastPersistTime     int64
	PersistInterval     time.Duration
	PersistTicker       time.Ticker
}

func (d *Database) Persist(fname string) (err error) {
	if d.LastPersistOpsCount == d.OpsCount {
		return
	}
	err = d.Db.DumpToFile(fname)
	if err != nil {
		err = errors.New(fmt.Sprintf("Could persist the db %s: %v", d.Name, err))
	} else {
		d.Lock()
		d.LastPersistOpsCount = d.OpsCount
		d.LastPersistTime = time.Now().UnixNano()
		d.Unlock()
	}
	return
}

func (d *Database) OpsLimitPersist(fname string) (err error) {
	if d.LastPersistOpsCount+d.PersistOpsLimit >= d.OpsCount {
		err = d.Persist(fname)
	}
	return
}

func (d *Database) Backup(srcFilePath string, dstPath string, dstFile string) (err error) {
	exists, err := tris.PathExists(dstPath)
	if !exists {
		if err != nil {
			err = errors.New(fmt.Sprintf("Could not stat directory %s for backup files: %v", dstPath, err))
			return
		} else {
			err = os.Mkdir(dstPath, 0777)
			if err != nil {
				err = errors.New(fmt.Sprintf("Could not create directory %s for backup files: %v", dstPath, err))
				return
			}
		}
	}

	// copy the old file into the backup folder
	// TODO: this should be dropped and be replaced by a write ops log + timestamp
	d.Db.Root.Lock()
	err = tris.CopyFile(srcFilePath, fmt.Sprintf("%s/%s", dstPath, dstFile))
	d.Db.Root.Unlock()
	if err != nil {
		err = errors.New(fmt.Sprintf("Could backup the previous data file: %v", err))
	}
	return
}
