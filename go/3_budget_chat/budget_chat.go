package main

import (
	"bufio"
	"bytes"
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
	firstJoinMessage = "* You are the first person to join the chatroom!\n"
)

var (
	alphanumericRegexp = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	ErrEmptyName       = errors.New("Name cannot be an empty string")
	ErrInvalidName     = errors.New("Name can only contain alphanumeric characters")
	ErrNameTooLong     = errors.New("Name cannot be longer than 16 characters")
)

type budgetChatServer struct {
	ln        net.Listener
	subch     chan *client
	unsubch   chan *client
	broadcast chan []byte
	clients   []*client
}

type client struct {
	name    string
	receive chan []byte
}

func main() {
	s := NewServer()
	s.serveBudgetChat()
}

func NewServer() budgetChatServer {
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Println("listen: ", err.Error())
		os.Exit(1)
	}
	log.Println("listening on port 9001")
	return budgetChatServer{
		ln,
		make(chan *client),
		make(chan *client),
		make(chan []byte),
		nil,
	}
}

func (b *budgetChatServer) serveBudgetChat() {
	go func() {
		for {
			select {
			case subscription := <-b.subch:
				if isDuplicate(subscription.name, b.clients) {
					subscription.receive <- []byte("duplicate")
				} else {
					subscription.receive <- []byte("ok")
					b.clients = append(b.clients, subscription)
					var msg string
					for _, e := range b.clients {
						if e == subscription {
							if len(b.clients) == 1 {
								msg = firstJoinMessage
							} else {
								msg = "* The room contains"
								for _, client := range b.clients[:len(b.clients)-1] {
									msg = msg + " " + client.name
								}
								msg = msg + "\n"
							}
						} else {
							msg = "* " + subscription.name + " has joined!\n"
						}
						log.Printf("Client list: %v", b.clients)
						log.Printf("Sending '%s' to %s", msg, e.name)
						e.receive <- []byte(msg)
					}
				}
			case unsub := <-b.unsubch:
				for i, e := range b.clients {
					if e == unsub {
						b.clients[i] = b.clients[len(b.clients)-1]
						b.clients = b.clients[:len(b.clients)-1]
					} else {
						e.receive <- []byte("* " + unsub.name + " has left!\n")
					}
				}
			case broadcast := <-b.broadcast:
				for _, e := range b.clients {
					name, _, _ := strings.Cut(string(broadcast), "]")
					if string(name[1:]) == e.name {
						continue
					} else {
						e.receive <- broadcast
					}
				}
			}
		}
	}()

	for {
		conn, err := b.ln.Accept()
		if err != nil {
			log.Println("accept: ", err.Error())
			os.Exit(1)
		}
		log.Println("connection from", conn.RemoteAddr())
		go handle(conn, b.unsubch, b.subch, b.broadcast)
	}
}

func handle(conn net.Conn, unsubch, subch chan *client, broadcast chan []byte) {
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
	name = name[:len(name)-1] // remove trailing \n
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
	}

	c := client{
		name:    string(name),
		receive: make(chan []byte),
	}
	prefix := []byte("[" + string(name) + "]" + " ")

	subch <- &c

	if bytes.Compare(<-c.receive, []byte("ok")) != 0 {
		_, err := conn.Write([]byte("That name is already taken.\n"))
		if err != nil {
			log.Println("Error writing to connection", err)
		}
		return
	}

	log.Printf("Registering client %v, as %s!", conn.RemoteAddr(), strings.TrimRight(string(name), "\n"))

	defer func() {
		unsubch <- &c
	}()

	go func() {
		for {
			inmsg := <-c.receive
			log.Printf("%s Recieved message '%s'", c.name, inmsg)
			_, err := conn.Write([]byte(inmsg))
			if err != nil {
				log.Println("Error writing to connection", err)
			}
		}
	}()

	for {
		outmsg, e := b.ReadBytes('\n')
		if e != nil {
			break
		}
		outmsg = append(prefix, outmsg...)
		broadcast <- outmsg
	}
}

func validateName(name string) error {
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

func isDuplicate(name string, clients []*client) bool {
	for _, c := range clients {
		if c.name == name {
			return true
		}
	}
	return false
}
