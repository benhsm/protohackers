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
