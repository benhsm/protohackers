package main

import (
	"bufio"
	"net"
	"testing"
)

func TestIntercept(t *testing.T) {
	s := NewMITMServer()
	go s.Serve()

	c, err := net.Dial("tcp", portString)
	if err != nil {
		t.Errorf("Could not connect to server: %v", err)
	}
	r := bufio.NewReader(c)
	r.ReadBytes('\n')
	c.Write([]byte("alice\n"))
	r.ReadBytes('\n')

	c2, err := net.Dial("tcp", portString)
	if err != nil {
		t.Errorf("could not connect to server: %v", err)
	}
	r2 := bufio.NewReader(c2)
	r2.ReadBytes('\n')
	c2.Write([]byte("bob\n"))
	r2.ReadBytes('\n')

	r.ReadBytes('\n')
	c2.Write([]byte("Hi alice, please send payment to " + address1 + "\n"))
	got, err := r.ReadBytes('\n')
	want := "[bob] Hi alice, please send payment to " + tonysAddress + "\n"
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

const (
	address1  = "7F1u3wSD5RbOHQmupo9nx4TnhQ"
	address2  = "7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX"
	address3  = "7LOrwbDlS8NujgjddyogWgIM93MV5N2VR"
	address4  = "7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T"
	lt26chars = "7F1u3wSD5RbOHQmupo9nx4Tnh"
	gt35chars = "7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8TA"
	no7       = "LOrwbDlS8NujgjddyogWgIM93MV5N2VR"
	nonAlnum  = "7LOrwbDl$8Nujgjd!yogWgIM93MV5N2VR"
)

func TestRewriteMsg(t *testing.T) {
	tt := []struct {
		test    string
		message string
		want    string
	}{
		{
			"address at end of message",
			"Hi alice, please send payment to " + address1,
			"Hi alice, please send payment to " + tonysAddress,
		},
		{
			"address at start of message",
			address4 + " is my boguscoin adress",
			tonysAddress + " is my boguscoin adress",
		},
		{
			"address in middle of message",
			"please send payment to " + address2 + " as soon as you can",
			"please send payment to " + tonysAddress + " as soon as you can",
		},
		{
			"address-like, but with less than 26 chars",
			"please send payment to " + lt26chars + " as soon as you can",
			"please send payment to " + lt26chars + " as soon as you can",
		},
		{
			"address-like, but with greater than 35 chars",
			"please send payment to " + gt35chars + " as soon as you can",
			"please send payment to " + gt35chars + " as soon as you can",
		},
		{
			"address-like, but does not start with 7",
			"please send payment to " + no7 + " as soon as you can",
			"please send payment to " + no7 + " as soon as you can",
		},
		{
			"address-like, but with non-alphanumeric characters",
			"please send payment to " + nonAlnum + " as soon as you can",
			"please send payment to " + nonAlnum + " as soon as you can",
		},
	}
	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			got := rewriteMsg([]byte(tc.message))
			if string(got) != tc.want {
				t.Errorf("got '%s', want '%s'", got, tc.want)
			}
		})
	}
}
