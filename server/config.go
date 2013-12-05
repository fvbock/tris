package tris

import (
	"log"
	"time"
)

type ServerConfig struct {
	Protocol string
	Host     string
	Port     int

	DataDir           string
	StorageFilePrefix string

	// persistance intervals on a per db basis
	PersistInterval time.Duration
	PersistOpsLimit int

	Logger *log.Logger
}
