package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

const (
	welcomeString    = "You've connected to budget chat. Please input a username to use for this session.\n"
	nameErrorMessage = "Usernames must contain only alphanumeric characters and cannot be longer than 16 characters in length.\n"
)

var (
	alphanumericRegexp = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	ErrEmptyName       = errors.New("Name cannot be an empty string")
	ErrInvalidName     = errors.New("Name can only contain alphanumeric characters")
	ErrNameTooLong     = errors.New("Name cannot be longer than 16 characters")
)

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
	defer log.Println("closed connection from", conn.RemoteAddr())
	b := bufio.NewReader(conn)
	conn.Write([]byte(welcomeString))
	log.Println("Sent welcome message to", conn.RemoteAddr())

	name, e := b.ReadBytes('\n')
	if e != nil {
		log.Printf("Error reading username from %v, %v", conn.RemoteAddr(), e)
		return
	}
	e = validateName(string(name))
	log.Printf("Recieved name %s", string(name))

	switch e {
	case ErrEmptyName:
		conn.Write([]byte("You need a username to continue.\n"))
		return
	case ErrInvalidName:
		conn.Write([]byte(nameErrorMessage))
		return
	case ErrNameTooLong:
		conn.Write([]byte(nameErrorMessage))
		return
	case nil:
		log.Printf("Registering client %v, as %s!", conn.RemoteAddr(), strings.TrimRight(string(name), "\n"))
	}

	for {
		_, e := b.ReadBytes('\n')
		if e != nil {
			break
		}
	}
}

func validateName(name string) error {
	name = strings.TrimRight(name, "\n")
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 16 {
		return ErrNameTooLong
	}
	if !alphanumericRegexp.Match([]byte(name)) {
		return ErrInvalidName
	}
	return nil
}
