package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

const welcomeString = "You've connected to budget chat. Please input a username to use for this session: \n"

func main() {
	s := StartServer()
	serveBudgetChat(s)
}

func StartServer() net.Listener {
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Println("listen: ", err.Error())
		os.Exit(1)
	}
	log.Println("listening on port 9001")
	return ln
}

func serveBudgetChat(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept: ", err.Error())
			os.Exit(1)
		}
		log.Println("connection from", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	b := bufio.NewReader(conn)
	conn.Write([]byte(welcomeString))
	log.Println("Sent welcome message to", conn.RemoteAddr())
	for {
		_, e := b.ReadBytes('\n')
		if e != nil {
			break
		}
	}
	log.Println("closed connection from", conn.RemoteAddr())
}
