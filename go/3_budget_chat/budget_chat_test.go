package main

import (
	"bufio"
	"bytes"
	"net"
	"strings"
	"testing"
	"time"
)

func TestBudgetChat(t *testing.T) {
	s := StartServer()
	go serveBudgetChat(s)

	var got string

	t.Run("sends a welcome message to incoming clients", func(t *testing.T) {
		conn := newConnection(t)
		defer conn.Close()
		got = readMessage(t, conn)
		assertStringEqual(t, got, welcomeString)
	})

	t.Run("prompts clients for name and announces names to clients", func(t *testing.T) {

		// User 1, Alice, connects
		conn1 := newConnection(t)
		defer conn1.Close()
		_ = readMessage(t, conn1)
		writeMessage(t, conn1, "Alice\n")
		got = readMessage(t, conn1)
		assertStringEqual(t, got, firstJoinMessage)

		// User 2, Bob, connects
		conn2 := newConnection(t)
		defer conn2.Close()
		_ = readMessage(t, conn2)
		writeMessage(t, conn2, "Bob\n")

		// Server tells Bob that Alice is here
		got = readMessage(t, conn2)
		if !strings.Contains(got, "Alice") ||
			got[0] != '*' {
			t.Errorf("got %s instead of expected control message", got)
		}

		// Server tells Alice that Bob joined
		got = readMessage(t, conn1)
		if !strings.Contains(got, "Bob") ||
			got[0] != '*' {
			t.Errorf("got %s instead of expected control message", got)
		}
	})

	t.Run("rejects invalid usernames", func(t *testing.T) {
		badConn := newConnection(t)
		defer badConn.Close()
		_ = readMessage(t, badConn)
		writeMessage(t, badConn, "#!)*(@#$!\n")
		got = readMessage(t, badConn)
		assertStringEqual(t, got, nameErrorMessage)

		badConn2 := newConnection(t)
		defer badConn2.Close()
		_ = readMessage(t, badConn2)
		writeMessage(t, badConn2, "thisismorethansixteencharacters\n")
		got = readMessage(t, badConn2)
		assertStringEqual(t, got, nameErrorMessage)

		badconn3 := newConnection(t)
		defer badconn3.Close()
		_ = readMessage(t, badconn3)
		writeMessage(t, badconn3, "You need a username to continue.\n")
		got = readMessage(t, badconn3)
		assertStringEqual(t, got, nameErrorMessage)
	})

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

func readMessage(t *testing.T, conn net.Conn) string {
	t.Helper()
	r := bufio.NewReader(conn)
	msg, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("error reading from connection: %v", err)
	}
	return string(msg)
}

func writeMessage(t *testing.T, conn net.Conn, msg string) {
	t.Helper()
	_, err := conn.Write([]byte(msg))
	if err != nil {
		t.Fatalf("error writing to connection: %v", err)
	}
}

func newConnection(t *testing.T) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", ":9001")
	conn.SetDeadline(time.Now().Add(time.Second * 5))
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

func assertStringEqual(t *testing.T, got, want string) {
	t.Helper()
	if want != got {
		t.Errorf("got %s, want %s", got, want)
	}
}
