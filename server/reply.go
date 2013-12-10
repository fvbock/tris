package tris

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fvbock/tris/util"
	"io"
	// "time"
)

const ()

type Reply struct {
	Payload [][]byte
	// Value    interface{}
	ReturnCode int64
	Type       int64
	Length     int64 // the nr of fields in one "reply row"
	Signature  []int // the fields in one "reply row"
}

func (r *Reply) Print() {
	if r.ReturnCode == COMMAND_OK {
		if len(r.Payload) == 1 {
			fmt.Printf("%s\n", r.Payload[0])
		} else {
			for i, rep := range r.Payload {
				fmt.Printf("%v) %s\n", i+1, rep)
			}
		}
		return
	}
	fmt.Printf("%s (Return code %v)\n", r.Payload[0], r.ReturnCode)
}

func (r *Reply) Serialize() (ser []byte) {
	rc := make([]byte, 1)
	_ = binary.PutVarint(rc, r.ReturnCode)
	ser = append(ser, rc...)
	rl := make([]byte, 1)
	_ = binary.PutVarint(rl, r.Length)
	ser = append(ser, rl...)
	for _, payload := range r.Payload {
		pl := make([]byte, 4)
		_ = binary.PutVarint(pl, int64(len(payload)))
		ser = append(ser, pl...)
		ser = append(ser, payload...)
	}

	return
}

func Unserialize(r []byte) *Reply {
	// var unserStart = time.Now()
	// defer func() { fmt.Printf("Unserialize took %v\n", time.Since(unserStart)) }()
	buf := bytes.NewReader(r)
	var rc int64
	rc, _, err := tris.ReadVarint(buf)
	if err != nil {
		fmt.Println(err)
	}
	rl, _, err := tris.ReadVarint(buf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("rc, rl", rc, rl)
	reply := &Reply{
		ReturnCode: rc,
		Length:     rl,
	}
	for {
		pLength, bl, err := tris.ReadVarint(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		}
		buf.Seek(int64(4-bl), 1)
		if pLength > 0 {
			payload := make([]byte, pLength)
			bRead, err := buf.Read(payload)
			if err != nil || int64(bRead) != pLength {
				fmt.Println("frakk. not enough bytes to read", err)
				break
			}
			reply.Payload = append(reply.Payload, payload)
		}
	}
	return reply
}

func NewReply(payload [][]byte, returnCode int64, responseLength int64) *Reply {
	return &Reply{
		Payload:    payload,
		ReturnCode: returnCode,
		Length:     responseLength,
	}
}

type MultiReply struct {
}
