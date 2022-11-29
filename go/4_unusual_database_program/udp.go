package main

import (
	"log"
	"net"
	"strings"
)

const (
	versionString = "Ben's UDP key-value store v0.1.0"
	portString    = ":4444"
)

func main() {
	s := NewServer()
	s.Serve()
}

type UDPServer struct {
	net.PacketConn
	kvstore map[string]string
	quit    chan bool
}

func NewServer() UDPServer {
	udplistener, err := net.ListenPacket("udp", portString)
	if err != nil {
		log.Fatal(err)
	}
	result := UDPServer{
		udplistener,
		make(map[string]string),
		make(chan bool),
	}
	result.kvstore["version"] = versionString
	return result
}

func (s *UDPServer) Serve() {
	select {
	case <-s.quit:
		return
	default:
		for {
			buf := make([]byte, 1000) // messages are a maximum of 1000 bytes
			_, addr, err := s.ReadFrom(buf)
			log.Printf("received a msg %s from %v", buf, addr)
			if err != nil {
				log.Printf("error reading from connection %v", err)
				continue
			}

			msg := strings.TrimRight(string(buf), "\000")
			key, value, found := strings.Cut(msg, "=")
			switch {
			case found:
				s.Insert(key, value)
			case !found:
				rsp := s.Retrieve(key)
				log.Printf("sending message %s to %v", rsp, addr)
				s.WriteTo(rsp, addr)
			}
		}
	}
}

func (s *UDPServer) Insert(key, value string) {
	// attempts to modify the version string are ignored
	if key == "version" {
		return
	}
	s.kvstore[key] = value
}

func (s *UDPServer) Retrieve(key string) []byte {
	rsp, _ := s.kvstore[key]
	rsp = key + "=" + rsp
	return []byte(rsp)
}
