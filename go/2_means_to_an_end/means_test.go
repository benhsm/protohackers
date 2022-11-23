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
	tt := []struct {
		test            string
		message         []byte
		wantMessageType int
		wantArg1        int32
		wantArg2        int32
		wantErr         error
	}{
		{
			"reads simple insert message",
			newMessage(insert, 12345, 101),
			insert,
			12345,
			101,
			nil,
		},
		{
			"reads simple query message",
			newMessage(query, 1000, 10000),
			query,
			1000,
			10000,
			nil,
		},
		{
			"reads insert message with negative price",
			newMessage(insert, 12345, -101),
			insert,
			12345,
			-101,
			nil,
		},
		{
			"returns an error if first byte does not specify type",
			[]byte{byte('A'), 0, 0, 48, 57, 0, 0, 0, 101},
			0,
			0,
			0,
			ErrNoMessageType,
		},
		{
			"returns an error if the message is less than 9 bytes",
			[]byte{byte('A'), 0, 0, 48, 57, 101},
			0,
			0,
			0,
			ErrMessageLength,
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			gotMessageType, gotArg1, gotArg2, err := readMessage(tc.message)
			if tc.wantErr != err {
				t.Errorf("Got unexpected error: %v, want %v", err, tc.wantErr)
			}
			if gotMessageType != tc.wantMessageType {
				t.Errorf("Got message type %v, want %v", gotMessageType, tc.wantMessageType)
			}
			if gotArg1 != tc.wantArg1 {
				t.Errorf("Got Arg1 %v, want %v", gotArg1, tc.wantArg1)
			}
			if gotArg2 != tc.wantArg2 {
				t.Errorf("Got Arg2 %v, want %v", gotArg2, tc.wantArg2)
			}
		})
	}
	t.Logf("%x", newMessage(insert, 12345, 101))
}

// helper functions

func newMessage(messageType int, arg1, arg2 int32) []byte {
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
