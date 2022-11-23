package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	fmt.Println("vim-go")
}

func StartServer() net.Listener {
	ln, err := net.Listen("tcp", ":1337")
	if err != nil {
		log.Println("listen: ", err.Error())
		os.Exit(1)
	}
	log.Println("Listening on port 1337")
	return ln
}

func ServeMeans(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept: ", err.Error())
			os.Exit(1)
		}
		log.Println("connection from ", conn.RemoteAddr())
		go Handle(conn)
	}
}

func Handle(conn net.Conn) {
	defer conn.Close()
	var message []byte
	bytesRead, err := conn.Read(message)
	if err != nil || bytesRead != 9 {
		conn.Write([]byte("Invalid request"))
		return
	}
}
