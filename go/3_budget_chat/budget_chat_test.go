package main

import (
	"bufio"
	"bytes"
	"net"
	"testing"
)

func TestBudgetChat(t *testing.T) {
	s := StartServer()
	go serveBudgetChat(s)

	conn1 := newConnection(t)
	defer conn1.Close()

	got := readMessage(t, conn1)
	assertBytesEqual(t, got, []byte(welcomeString))

	writeMessage(t, conn1, []byte("Alice\n"))
	got = readMessage(t, conn1)
	assertBytesEqual(t, got, []byte(firstJoinMessage))
}

func TestValidateName(t *testing.T) {
	tt := []struct {
		test string
		name string
		want error
	}{
		{
			"accepts a name containing letters",
			"Alice\n",
			nil,
		},
		{
			"accepts a name containing alphanumeric characters",
			"Bob123\n",
			nil,
		},
		{
			"rejects names containing non-alphanumeric characters",
			"Chuck!\n",
			ErrInvalidName,
		},
		{
			"rejects names which are just empty strings",
			"\n",
			ErrEmptyName,
		},
		{
			"rejects names longer than 16 characters",
			"12345678901234567\n",
			ErrNameTooLong,
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			got := validateName(tc.name)
			if got != tc.want {
				t.Errorf("expected err '%v', got '%v'", tc.want, got)
			}
		})
	}
}

// helper functions

func readMessage(t *testing.T, conn net.Conn) []byte {
	t.Helper()
	r := bufio.NewReader(conn)
	msg, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("error reading from connection: %v", err)
	}
	return msg
}

func writeMessage(t *testing.T, conn net.Conn, msg []byte) {
	t.Helper()
	_, err := conn.Write(msg)
	if err != nil {
		t.Fatalf("error writing to connection: %v", err)
	}
}

func newConnection(t *testing.T) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", ":9001")
	if err != nil {
		t.Fatalf("error establishing connection: %v", err)
	}
	return conn
}

func assertBytesEqual(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Compare(got, want) != 0 {
		t.Errorf("got %s, want %s", got, want)
	}
}
