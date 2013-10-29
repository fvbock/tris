package main

import (
	"flag"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	trisclient "github.com/fvbock/tris/client"
	trisserver "github.com/fvbock/tris/server"
	"github.com/sbinet/liner"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

var term *liner.State = nil

var dsnString = flag.String("d", "tcp:localhost:6000", "dsn to connect to")

// var port = flag.String("p", "", "file to run")

func init() {
	fmt.Println(`TriS cli 0.0.1.
`)
	term = liner.NewLiner()

	fname := path.Join(os.Getenv("HOME"), ".tris.history")
	f, err := os.Open(fname)
	if err != nil {
		fmt.Printf("**warning: could not access history file [%s]\n", fname)
		return
	}
	defer f.Close()
	_, err = term.ReadHistory(f)
	if err != nil {
		fmt.Printf("**warning: could not read history file [%s]\n", fname)
		return
	}
}

func atexit() {
	fname := path.Join(os.Getenv("HOME"), ".go.history")
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("**warning: could not access history file [%s]\n", fname)
		return
	}
	defer f.Close()
	_, err = term.WriteHistory(f)
	if err != nil {
		fmt.Printf("**warning: could not write history file [%s]\n", fname)
		return
	}

	err = term.Close()
	if err != nil {
		fmt.Printf("**warning: problem closing term: %v\n", err)
		return
	}
	fmt.Println("\n")
}

func main() {
	defer atexit()

	var ierr error = nil // previous interpreter error
	ps1 := "tris [not connected]> "
	// ps2 := "...  "
	prompt := &ps1
	command := ""

	// TRIS conn
	flag.Parse()
	dsnParts := strings.Split(*dsnString, ":")
	fmt.Printf("Connecting to %s://%s:%s\n", dsnParts[0], dsnParts[1], dsnParts[2])

	port, _ := strconv.ParseInt(dsnParts[2], 10, 32)
	dsn := &trisclient.DSN{
		Protocol: dsnParts[0],
		Host:     dsnParts[1],
		Port:     int(port),
	}
	fmt.Printf("Connecting to %s://%s:%v\n", dsn.Protocol, dsn.Host, dsn.Port)
	dsn = &trisclient.DSN{
		Protocol: "tcp",
		Host:     "localhost",
		Port:     6000,
	}
	ctx, err := zmq.NewContext()
	if err != nil {
		fmt.Println("Context error:", ierr)
	}
	client, err := trisclient.NewClient(dsn, ctx)
	err = client.Dial()
	if err != nil {
		fmt.Println(command, ierr)
	} else {
		ps1 = fmt.Sprintf("tris [%s:%v]> ", dsnParts[1], dsnParts[2])
	}

	for {
		line, err := term.Prompt(*prompt)
		if err != nil {
			if err != io.EOF {
				ierr = err
			} else {
				ierr = nil
			}
			break //os.Exit(0)
		}
		if line == "" || line == ";" {
			// no more input
			prompt = &ps1
		}

		command += line
		if command != "" {
			for _, ll := range strings.Split(command, "\n") {
				term.AppendHistory(ll)
			}
		} else {
			continue
		}

		command = command + "\n"
		// fmt.Println(command, ierr)
		r, err := client.Send(command)
		if err != nil {
			fmt.Println("Error:", err)
		}

		// fmt.Println(string(r))
		response := trisserver.Unserialize(r)
		fmt.Println(response)

		//	resetstate:
		// reset state
		command = ""
		ierr = nil
	}

}