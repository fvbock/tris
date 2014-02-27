package tris

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fvbock/tris/util"
	"io"
	// "time"
	"strings"
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

// func (r *Reply) Rows() []*ReplyData {
// 	if r.ReturnCode == COMMAND_OK {
// 	}
// }

type ReplyData struct {
	Key   string
	Count int64
	Data  interface{}
}

func (r *Reply) Print() {
	if r.ReturnCode == COMMAND_OK {
		var n int = 1
		var iLen int
		for rCount, pItem := range r.Payload {
			// fmt.Println(">>>", rCount, pItem)
			var row string
			rType := r.Signature[rCount%len(r.Signature)]
			// for n, rType := range r.Signature {
			switch rType {
			case REPLY_TYPE_BOOL:
				buf := bytes.NewReader(pItem)
				data, err := decodeBoolReply(buf)
				if err != nil {
					fmt.Println("ERROR: decoding failed:", err)
				}
				row = fmt.Sprintf("%v", data)
			case REPLY_TYPE_INT:
				buf := bytes.NewReader(pItem)
				data, err := decodeIntReply(buf)
				if err != nil {
					fmt.Println("ERROR: decoding failed:", err)
				}
				row = fmt.Sprintf("%v", data)
			case REPLY_TYPE_STRING:
				// data, err := decodeStringReply(buf)
				// if err != nil {
				// 	fmt.Println("ERROR: decoding failed:", err)
				// }
				row = fmt.Sprintf("%s", pItem)
			default:
				fmt.Println("ERROR: got unknown response type:", rType)
			}

			if len(r.Signature) == 1 && r.Signature[0] == REPLY_TYPE_STRING {
				fmt.Printf("%s\n", row)
			} else {
				if rCount%len(r.Signature) == 0 {
					fmt.Printf("%v) %s\n", n, row)
					iLen = len(fmt.Sprintf("%v", n))
					n++
				} else {
					fmt.Printf("%s%s\n", strings.Repeat(" ", iLen+2), row)
				}
			}
		}

		return
	}
	fmt.Printf("%s (Return code %v)\n", r.Payload[0], r.ReturnCode)
}

func (r *Reply) Serialize() (ser []byte) {
	// the return code
	rc := make([]byte, 1)
	_ = binary.PutVarint(rc, r.ReturnCode)
	ser = append(ser, rc...)

	// response item length (fields per reply item)
	rl := make([]byte, 1)
	// _ = binary.PutVarint(rl, r.Length)
	_ = binary.PutVarint(rl, int64(len(r.Signature)))
	ser = append(ser, rl...)

	// response signature
	// fmt.Println("SIG", r.Signature)
	for _, rType := range r.Signature {
		rt := make([]byte, 1)
		_ = binary.PutVarint(rt, int64(rType))
		ser = append(ser, rt...)
	}

	// serializing the payload(s) - the single reply items

	// for _, payload := range r.Payload {
	// 	// byte length of the item
	// 	pl := make([]byte, 4)
	// 	_ = binary.PutVarint(pl, int64(len(payload)))
	// 	ser = append(ser, pl...)
	// 	// adding the items data
	// 	fmt.Println("@@@", payload, string(payload))
	// 	ser = append(ser, payload...)
	// }

	for idx, payload := range r.Payload {
		// fmt.Println("***", idx, payload, string(payload))
		if len(r.Signature) == 0 {
			pl := make([]byte, 4)
			_ = binary.PutVarint(pl, int64(0))
			ser = append(ser, pl...)
			continue
		}
		rtype := r.Signature[idx%len(r.Signature)]
		switch rtype {
		case REPLY_TYPE_BOOL:
			// TODO: dont use strings
			bData := make([]byte, 4)
			if string(payload) == "TRUE" {
				_ = binary.PutVarint(bData, int64(1))
			} else {
				_ = binary.PutVarint(bData, int64(0))
			}
			ser = append(ser, bData...)
		case REPLY_TYPE_INT:
			ser = append(ser, payload...)
		case REPLY_TYPE_STRING:
			ser = append(ser, encodeStringReply(payload)...)
		default:
			fmt.Println("ERROR: got unknown response type:", rtype)
		}
	}

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

func decodeBoolReply(reply *bytes.Reader) (r bool, err error) {
	rval, _, err := tris.ReadVarint(reply)
	if err != nil {
		fmt.Println(err)
		return
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

func decodeIntReply(reply *bytes.Reader) (r int64, err error) {
	r, _, err = tris.ReadVarint(reply)
	if err != nil {
		if err == io.EOF {
			fmt.Println(err)
			return
		}
	}
	// reply.Seek(int64(4-bl), 1)

	return
}

func encodeStringReply(r []byte) (sr []byte) {
	// byte length of the item
	pl := make([]byte, 4)
	_ = binary.PutVarint(pl, int64(len(r)))
	sr = append(sr, pl...)
	// adding the items data
	sr = append(sr, r...)
	return
}

func decodeStringReply(reply *bytes.Reader) (r []byte, err error) {
	pLength, bl, err := tris.ReadVarint(reply)
	if err != nil {
		if err == io.EOF {
			fmt.Println(err)
			return
		}
	}
	reply.Seek(int64(4-bl), 1)

	if pLength > 0 {
		r = make([]byte, pLength)
		bRead, err := reply.Read(r)
		if err != nil || int64(bRead) != pLength {
			fmt.Println("frakk. not enough bytes to read", err)
		}
	}

	return
}

// func encodeFloatReply(r int64) {
// }
// func decodeFloatReply(r int64) {
// }

func Unserialize(r []byte) *Reply {
	// var unserStart = time.Now()
	// defer func() { fmt.Printf("Unserialize took %v\n", time.Since(unserStart)) }()
	// fmt.Println("raw response", r)
	buf := bytes.NewReader(r)
	var rc int64

	// read the return code
	rc, _, err := tris.ReadVarint(buf)
	if err != nil {
		fmt.Println(err)
	}

	// read response field length
	rl, _, err := tris.ReadVarint(buf)
	if err != nil {
		fmt.Println(err)
	}

	// read response signature
	rSig := []int{}
	for i := 0; i < int(rl); i++ {

		fieldType, _, err := tris.ReadVarint(buf)
		if err != nil {
			fmt.Println(err)
		}
		rSig = append(rSig, int(fieldType))
	}

	// fmt.Println("rc, rl", rc, rl)
	reply := &Reply{
		ReturnCode: rc,
		Length:     rl,
		Signature:  rSig,
	}
	// fmt.Println("SIG", reply.Signature)

readLoop:
	for {
		if len(reply.Signature) == 0 {
			break
		}
		for _, rType := range reply.Signature {
			if rType != REPLY_TYPE_STRING {
				payload := make([]byte, 4)
				_, err := buf.Read(payload)
				if err != nil {
					if err == io.EOF {
						break readLoop
					}
					fmt.Println(err)
				}
				reply.Payload = append(reply.Payload, payload)
			} else {
				pLength, bl, err := tris.ReadVarint(buf)
				if err != nil {
					if err == io.EOF {
						break readLoop
					}
					fmt.Println(err)
				}
				buf.Seek(int64(4-bl), 1)
				if pLength > 0 {
					payload := make([]byte, pLength)
					bRead, err := buf.Read(payload)
					if err != nil || int64(bRead) != pLength {
						fmt.Println("frakk. not enough bytes to read", err)
						break readLoop
					}
					reply.Payload = append(reply.Payload, payload)
				}
			}
		}
	}

	// fmt.Println("&&&", reply)
	return reply
}

func NewReply(payload [][]byte, returnCode int64, responseLength int64, signature []int) *Reply {
	return &Reply{
		Payload:    payload,
		ReturnCode: returnCode,
		Length:     responseLength,
		Signature:  signature,
	}
}

type MultiReply struct {
}
