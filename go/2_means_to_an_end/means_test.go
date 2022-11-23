package main

import (
	"bytes"
	"net"
	"testing"
)

func TestReadMessage(t *testing.T) {
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
