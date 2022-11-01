package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func Test_validateRequest(t *testing.T) {
	tt := []struct {
		test    string
		payload []byte
		want    bool
	}{
		{
			"Simple valid request",
			[]byte(`{"method":"isPrime","number":123}`),
			true,
		},
		{
			"Invalid request: Not JSON",
			[]byte(`Hello world!`),
			false,
		},
		{
			"Invalid request: missing required field 'method'",
			[]byte(`{"haha":"isPrime","number":123}`),
			false,
		},
		{
			"Invalid request: missing required field 'number'",
			[]byte(`{"method":"isPrime","haha":"haha"}`),
			false,
		},
		{
			"Invalid request: method other than 'isPrime'",
			[]byte(`{"method":"bestMethod","number":123}`),
			false,
		},
		{
			"Invalid request: number field contains non-number value",
			[]byte(`{"method":"isPrime","number":"onetwothree"}`),
			false,
		},
		{
			"Valid request with extraneous fields",
			[]byte(`{"method":"isPrime","number":123, "haha":"isFunny"}`),
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			validity, _ := validateRequest(tc.payload)
			if validity != tc.want {
				t.Errorf("Didn't get %v, with payload %s.", tc.want, tc.payload)
			}
		})
	}
}

func Test_isPrime(t *testing.T) {
	tt := []struct {
		test string
		num  int64
		want bool
	}{
		{
			"1 is not prime",
			1,
			false,
		},
		{
			"2 is prime",
			2,
			true,
		},
		{
			"3 is prime",
			2,
			true,
		},
		{
			"0 is not prime",
			0,
			false,
		},
		{
			"negative numbers are not prime",
			-10,
			false,
		},
		{
			"negative numbers are not prime",
			-7,
			false,
		},
		{
			"7 is prime",
			7,
			true,
		},
		{
			"8 is not prime",
			8,
			false,
		},
		{
			"7901 is prime",
			7901,
			true,
		},
		{
			"7902 is not prime",
			7902,
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			if isPrime(tc.num) != tc.want {
				t.Errorf("Didn't get %v, with number %d.", tc.want, tc.num)
			}
		})
	}
}

func Test_primeTime(t *testing.T) {
	ln, err := net.Listen("tcp", ":10000")
	if err != nil {
		fmt.Println("listen: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("listening on port 10000")
	go servePrimeTime(ln)

	tt := []struct {
		test    string
		payload []byte
		want    []byte
	}{
		{
			"Simple valid request, non prime number",
			[]byte(`{"method":"isPrime","number":123}` + "\n"),
			[]byte(`{"method":"isPrime","prime":false}` + "\n"),
		},
		{
			"Simple valid request, prime number",
			[]byte(`{"method":"isPrime","number":7}` + "\n"),
			[]byte(`{"method":"isPrime","prime":true}` + "\n"),
		},
		{
			"Invalid request: Not JSON",
			[]byte(`Hello world!` + "\n"),
			[]byte("Malformed Request." + "\n"),
		},
		{
			"Invalid request: missing required field 'method'",
			[]byte(`{"haha":"isPrime","number":123}` + "\n"),
			[]byte("Malformed Request." + "\n"),
		},
		{
			"Invalid request: missing required field 'number'",
			[]byte(`{"method":"isPrime","haha":"haha"}` + "\n"),
			[]byte("Malformed Request." + "\n"),
		},
		{
			"Invalid request: method other than 'isPrime'",
			[]byte(`{"method":"bestMethod","number":123}` + "\n"),
			[]byte("Malformed Request." + "\n"),
		},
		{
			"Invalid request: number field contains non-number value",
			[]byte(`{"method":"isPrime","number":"onetwothree"}` + "\n"),
			[]byte("Malformed Request." + "\n"),
		},
		{
			"Valid request with extraneous fields, nonprime",
			[]byte(`{"method":"isPrime","number":123, "haha":"isFunny"}` + "\n"),
			[]byte(`{"method":"isPrime","prime":false}` + "\n"),
		},
		{
			"Valid request with extraneous fields, prime",
			[]byte(`{"method":"isPrime","number":7, "haha":"isFunny", "extraField":1}` + "\n"),
			[]byte(`{"method":"isPrime","prime":true}` + "\n"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			conn, err := net.Dial("tcp", ":10000")
			conn.SetReadDeadline((time.Now().Add(2 * time.Second)))
			if err != nil {
				t.Error("could not connect to TCP server: ", err)
			}
			defer conn.Close()

			if _, err := conn.Write(tc.payload); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 1024)
			if _, err := conn.Read(out); err == nil {
				if !bytes.Equal(bytes.TrimRight(out, "\u0000"), tc.want) {
					t.Errorf("response did not match expected output. Got %s, wanted %s\n", bytes.TrimRight(out, "\u0000"), tc.want)
				} else {
					t.Log("got expected response")
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}

}
