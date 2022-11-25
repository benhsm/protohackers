package main

import (
	"bufio"
	"net"
	"testing"
)

func TestBudgetChat(t *testing.T) {
	s := StartServer()
	go serveBudgetChat(s)

	conn1, err := net.Dial("tcp", ":9001")
	if err != nil {
		t.Fatalf("could not connect to server: %v", err)
	}
	defer conn1.Close()

	r1 := bufio.NewReader(conn1)
	got, err := r1.ReadBytes('\n')
	if err != nil {
		t.Fatalf("error reading from connection: %v", err)
	}

	if string(got) != welcomeString {
		t.Errorf("wanted welcome message '%s', got '%s'", welcomeString, got)
	}
}
