package tris

import (
	"errors"
	"fmt"
	"time"
)

type TrisConnectionPool struct {
	Dsn      *DSN
	Pool     chan *Client
	PoolSize int
}

func NewTrisConnectionPool(dsn *DSN, poolSize int) (p *TrisConnectionPool, err error) {
	var poolStart = time.Now()
	defer func() { fmt.Printf("Pool initializatopn took %v\n", time.Since(poolStart)) }()
	p = &TrisConnectionPool{
		Pool:     make(chan *Client, poolSize),
		PoolSize: poolSize,
	}
	for i := 0; i < p.PoolSize; i++ {
		c, err := NewClient(dsn)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to initialize pool Client: %v.", err))
			break
		}
		err = c.Dial()
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to initialize pool Client: %v.", err))
			break
		}
		p.Pool <- c
	}
	return
}

func (p *TrisConnectionPool) Get() (c *Client, err error) {
	c = <-p.Pool
	return
}

func (p *TrisConnectionPool) Put(c *Client) (err error) {
	// TODO: name needs to go into some "common" package and be a constant there - not in the server
	c.Select("0")
	p.Pool <- c
	return
}

func (p *TrisConnectionPool) recycleClient(c *Client) (err error) {
	return
}
