package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"regexp"
	"unicode/utf8"
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

func rewriteMsg(src []byte) []byte {
	re := bogusCoinRE
	lastMatchEnd := 0 // end position of the most recent match
	searchPos := 0    // position where we next look for a match
	var buf []byte
	s := src
	endPos := len(s)

	for searchPos <= endPos {
		s = s[searchPos:]
		loc := re.FindIndex(s)
		if loc == nil {
			break // no more matches
		}

		// copy the unmatched characters before this match
		buf = append(buf, s[lastMatchEnd:loc[0]]...)

		// insert a copy of the replacement string
		if loc[1] > lastMatchEnd || loc[0] == 0 {
			buf = append(buf, []byte(tonysAddress)...)
		}
		lastMatchEnd = loc[1]

		// Advance past this match; always advance at least one character.
		var width int
		_, width = utf8.DecodeRune(s[searchPos:])

		if searchPos+width > loc[1] {
			searchPos += width
		} else if searchPos+1 > loc[1] {
			searchPos++
		} else {
			searchPos = loc[1]
		}
	}

	return append(buf, s...)
}
