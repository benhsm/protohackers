package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	insert = iota
	query
)

var (
	ErrNoMessageType = errors.New("message does not have a type specifier")
	ErrMessageLength = errors.New("message is less than 9 bytes")
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

func readMessage(message []byte) (int, int32, int32, error) {

	if len(message) < 9 {
		return 0, 0, 0, ErrMessageLength
	}

	var messageType int
	switch message[0] {
	case byte('I'):
		messageType = insert
	case byte('Q'):
		messageType = query
	default:
		return 0, 0, 0, ErrNoMessageType
	}

	arg1 := int32(binary.BigEndian.Uint32(message[1:5]))
	arg2 := int32(binary.BigEndian.Uint32(message[5:9]))

	//	return messageType, 12345, 101, nil
	return messageType, arg1, arg2, nil
}
