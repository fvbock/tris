package tris

import (
	"log"
)

type ServerConfig struct {
	Protocol string
	Host     string
	Port     int

	DataDir           string
	StorageFilePrefix string

	Logger *log.Logger
}
