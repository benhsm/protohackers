package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
)

func TestSmoke(t *testing.T) {
	ln, err := net.Listen("tcp", ":10000")
	if err != nil {
		fmt.Println("listen: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("listening on port 10000")
	go serveEcho(ln)

	tt := []struct {
		test    string
		payload []byte
		want    []byte
	}{
		{
			"Sending a simple request returns result",
			[]byte("hello world\n"),
			[]byte("hello world\n")},
		{
			"Sending another simple request works",
			[]byte("goodbye world\n"),
			[]byte("goodbye world\n"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			conn, err := net.Dial("tcp", ":10000")
			if err != nil {
				t.Error("could not connect to TCP server: ", err)
			}
			defer conn.Close()

			if _, err := conn.Write(tc.payload); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 1024)
			if _, err := conn.Read(out); err == nil {
				if bytes.Compare(out, tc.want) == 0 {
					t.Error("response did match expected output")
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}
}
