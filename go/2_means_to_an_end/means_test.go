package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
)

func TestHandleMessage(t *testing.T) {
	s := StartServer()
	go ServeMeans(s)

	tt := []struct {
		test    string
		message []byte
		want    []byte
	}{
		{
			"Invalid query",
			[]byte("hello server"),
			[]byte("Invalid request"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			conn, err := net.Dial("tcp", ":1337")
			if err != nil {
				t.Error("could not connect to TCP server: ", err)
			}
			defer conn.Close()

			if _, err := conn.Write(tc.message); err != nil {
				t.Error("could not write message to TCP server:", err)
			}

			out := make([]byte, 20)
			if _, err := conn.Read(out); err == nil {
				out = bytes.TrimRight(out, "\000")
				if !bytes.Equal(out, tc.want) {
					t.Errorf("got %v, wanted %v\n", out, tc.want)
				} else {
					t.Log("got expected response")
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}
}

func TestReadMessage(t *testing.T) {
	t.Logf("%x", newMessage(insert, 12345, 101))
}

// helper functions

const (
	insert = iota
	query
)

func newMessage(messageType uint8, arg1, arg2 int32) []byte {
	result := make([]byte, 1)
	if messageType == insert {
		result[0] = byte('I')
	} else if messageType == query {
		result[0] = byte('Q')
	}

	result = binary.BigEndian.AppendUint32(result, uint32(arg1))
	result = binary.BigEndian.AppendUint32(result, uint32(arg2))

	return result
}
