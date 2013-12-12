package main

import (
	"flag"
	"fmt"
	// zmq "github.com/alecthomas/gozmq"
	trisclient "github.com/fvbock/tris/client"
	trisserver "github.com/fvbock/tris/server"
	"github.com/sbinet/liner"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

// TODO: detect connection loss and handle reconnect + currentDB

var (
	term      *liner.State = nil
	dsnString              = flag.String("d", "tcp:localhost:6000", "dsn to connect to")
)

func init() {
	fmt.Printf("TriS cli %s\n", trisclient.VERSION)
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
	fmt.Println("CLI atexit")
	fname := path.Join(os.Getenv("HOME"), ".tris.history")
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
	ps1 := "[not connected]> "
	// ps2 := "...  "
	prompt := &ps1
	command := ""

	// TRIS conn
	flag.Parse()
	dsnParts := strings.Split(*dsnString, ":")
	port, _ := strconv.ParseInt(dsnParts[2], 10, 32)
	dsn := &trisclient.DSN{
		Protocol: dsnParts[0],
		Host:     dsnParts[1],
		Port:     int(port),
	}
	fmt.Printf("Connecting to %s://%s:%v\n", dsn.Protocol, dsn.Host, dsn.Port)

	client, err := trisclient.NewClient(dsn)
	err = client.Dial()
	if err != nil {
		fmt.Println(command, ierr)
	} else {
		ps1 = fmt.Sprintf("%s:%v/[%s]> ", dsnParts[1], dsnParts[2], client.ActiveDb)
	}
	defer client.Close()

	// check conn
	// r, err := client.Send("PING")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	response, err := client.Ping()
	// response := trisserver.Unserialize(r)
	if response.ReturnCode != trisserver.COMMAND_OK {
		fmt.Printf("Initial PING failed:\n%v\n", response)
	} else {
		for {
			ps1 = fmt.Sprintf("%s:%v/[%s]> ", dsnParts[1], dsnParts[2], client.ActiveDb)
			// fmt.Println(">>>", client.ActiveDb)
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

			// // nama
			// r, err := client.Send(command)
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// }

			// response := trisserver.Unserialize(r)
			// // fmt.Println(response)
			// response.Print()

			// use lib
			var cmds []string
			var args [][]string
			msgs := strings.Split(strings.Trim(command, " "), "\n")
			fmt.Println("msgs", msgs)
			for n, msg := range msgs {
				parts := strings.Split(strings.Trim(msg, " "), " ")
				for i, p := range parts {
					if len(p) == 0 {
						continue
					}
					if i == 0 {
						cmds = append(cmds, strings.ToUpper(p))
						args = append(args, make([]string, 0))
					} else {
						args[n] = append(args[n], p)
					}
				}
			}
			fmt.Println("cmds, args", cmds, args)

		cmdexec:
			for i, cmdname := range cmds {
				var response *trisserver.Reply
				var err error
				switch cmdname {
				case "PING":
					response, err = client.Ping()
				case "SELECT":
					response, err = client.Select(args[i][0])
				case "DBINFO":
					response, err = client.DbInfo()
				case "INFO":
					response, err = client.Info()
				case "SAVE":
					response, err = client.Save()
				case "IMPORT":
					response, err = client.ImportDb(args[i][0], args[i][1])
				case "MERGE":
					response, err = client.MergeDb(args[i][0])
				case "CREATE":
					response, err = client.Create(args[i][0])
				case "ADD":
					response, err = client.Add(args[i][0])
				case "DEL":
					response, err = client.Del(args[i][0])
				case "HAS":
					response, err = client.Has(args[i][0])
				case "HASCOUNT":
					response, err = client.HasCount(args[i][0])
				case "HASPREFIX":
					response, err = client.HasPrefix(args[i][0])
				case "MEMBERS":
					response, err = client.Members()
				case "PREFIXMEMBERS":
					response, err = client.PrefixMembers(args[i][0])
				default:
					fmt.Println("Unknown command.")
					break cmdexec
				}
				if response.ReturnCode != trisserver.COMMAND_OK || err != nil {
					// TODO
				}
				response.Print()
			}

			// reset state
			command = ""
			ierr = nil
		}
	}
}
