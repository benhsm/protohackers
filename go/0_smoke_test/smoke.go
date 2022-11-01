package main

// This is mostly the same as the example given at the bottom of:
// https://protohackers.com/help

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	setup()
}

func setup() {
	ln, err := net.Listen("tcp", ":10000")
	if err != nil {
		fmt.Println("listen: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("listening on port 10000")
	serveEcho(ln)
}

func serveEcho(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("connection from ", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	if _, err := io.Copy(conn, conn); err != nil {
		fmt.Println("copy: ", err.Error())
	}
}
