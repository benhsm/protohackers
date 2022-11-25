package main

import (
	"encoding/binary"
	"errors"
	"io"
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

type priceRecord struct {
	timestamp int32
	price     int32
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
	//	conn.SetDeadline(time.Now().Add(10 * time.Second))
	message := make([]byte, 9)
	var records []priceRecord
	for {
		_, err := io.ReadFull(conn, message)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error reading from connection: ", err)
			break
		}
		messageType, arg1, arg2, readErr := readMessage(message)
		if readErr != nil {
			log.Println("Invalid request: ", readErr)
			conn.Write([]byte("Invalid request"))
			break
		}
		switch messageType {
		case insert:
			if arg1 < 0 {
				conn.Write([]byte("Invalid request"))
				break
			}
			log.Printf("connection %s, I time %d price %d", conn.RemoteAddr().String(), arg1, arg2)
			records = append(records, priceRecord{arg1, arg2})
		case query:
			average, err := queryPriceRecord(records, arg1, arg2)
			if err != nil {
				conn.Write([]byte("Invalid request"))
				break
			}
			log.Printf("connection %s, Q min %d max %d", conn.RemoteAddr().String(), arg1, arg2)
			response := make([]byte, 4)
			binary.BigEndian.PutUint32(response, uint32(average))
			conn.Write(response)
		}
	}
	log.Println("Session ended: ", conn.RemoteAddr())
}

func queryPriceRecord(records []priceRecord, mintime, maxtime int32) (int32, error) {

	var sum int64
	var numRecords int32
	for _, record := range records {
		if mintime <= record.timestamp && record.timestamp <= maxtime {
			sum += int64(record.price)
			numRecords++
		}
	}

	if numRecords == 0 {
		return 0, nil
	}

	log.Printf("Sum of %d dividing by %d", sum, numRecords)

	return int32(sum / int64(numRecords)), nil
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
