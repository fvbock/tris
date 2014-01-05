package tris

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fvbock/tris/util"
	"io"
	// "time"
)

const (
	REPLY_TYPE_BOOL   = 0
	REPLY_TYPE_INT    = 1
	REPLY_TYPE_STRING = 2
	// REPLY_TYPE_FLOAT  = 3
)

type Reply struct {
	Payload [][]byte
	// Value    interface{}
	ReturnCode int64
	Type       int64
	Length     int64 // the nr of fields in one "reply row"
	Signature  []int // the fields in one "reply row"
}

type ReplyData struct {
	Key   string
	Count int64
	Data  interface{}
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
	// first the return code
	rc := make([]byte, 1)
	_ = binary.PutVarint(rc, r.ReturnCode)
	ser = append(ser, rc...)
	// response length (fields per reply item)
	rl := make([]byte, 1)
	// _ = binary.PutVarint(rl, r.Length)
	_ = binary.PutVarint(rl, int64(len(r.Signature)))
	ser = append(ser, rl...)
	// serializing the payload(s) - the single reply items
	for _, payload := range r.Payload {
		// byte length of the item
		pl := make([]byte, 4)
		_ = binary.PutVarint(pl, int64(len(payload)))
		ser = append(ser, pl...)
		// adding the items data
		ser = append(ser, payload...)
	}
	// for idx, payload := range r.Payload {
	// 	rtype := r.Signature[idx%len(r.Signature)]
	// 	switch rtype {
	// 	case REPLY_TYPE_BOOL:
	// 		// TODO: dont use strings
	// 		if string(payload) == "TRUE" {
	// 			_ = binary.PutVarint(pl, int64(1))
	// 		} else {
	// 			_ = binary.PutVarint(pl, int64(0))
	// 		}
	// 	case REPLY_TYPE_INT:
	// 		pl := make([]byte, 4)
	// 		_ = binary.PutVarint(pl, int64(len(payload)))
	// 	case REPLY_TYPE_STRING:
	// 		// byte length of the item
	// 		pl := make([]byte, 4)
	// 		_ = binary.PutVarint(pl, int64(len(payload)))
	// 		ser = append(ser, pl...)
	// 		// adding the items data
	// 		ser = append(ser, payload...)
	// 	}
	// }

	return
}

func encodeBoolReply(r bool) (br []byte) {
	if r {
		br := make([]byte, 1)
		_ = binary.PutVarint(br, int64(1))
	} else {
		_ = binary.PutVarint(br, int64(0))
	}
	return
}

func decodeBoolReply(reply *bytes.Reader) (r bool) {
	rval, _, err := tris.ReadVarint(reply)
	if err != nil {
		fmt.Println(err)
	}
	if rval == 1 {
		r = true
	} else {
		r = false
	}
	return
}

func encodeIntReply(r int64) (ir []byte) {
	ir = make([]byte, 4)
	_ = binary.PutVarint(ir, r)
	return
}

func decodeIntReply(reply *bytes.Reader) (r int64) {
	r, bl, err := tris.ReadVarint(reply)
	if err != nil {
		if err == io.EOF {

		}
		fmt.Println(err)
	}
	reply.Seek(int64(4-bl), 1)

	return
}

func encodeStringReply(r string) (sr []byte) {
	// byte length of the item
	pl := make([]byte, 4)
	_ = binary.PutVarint(pl, int64(len(r)))
	sr = append(sr, pl...)
	// adding the items data
	sr = append(sr, r...)
	return
}

// func encodeFloatReply(r int64) {
// }
// func decodeFloatReply(r int64) {
// }

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
	// fmt.Println("rc, rl", rc, rl)
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
