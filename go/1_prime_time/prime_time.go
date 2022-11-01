package main

import (
	"bufio"
	"encoding/json"
	"log"
	"math"
	"net"
	"os"
)

func main() {
	ln, err := net.Listen("tcp", ":1337")
	if err != nil {
		log.Println("listen: ", err.Error())
		os.Exit(1)
	}

	servePrimeTime(ln)
}

func servePrimeTime(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept: ", err.Error())
			os.Exit(1)
		}
		log.Println("connection from ", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	b := bufio.NewReader(conn)
	for {
		request, e := b.ReadBytes('\n')
		if e != nil {
			break
		}
		log.Printf("Recieved request: %s", request)

		valid, num := validateRequest(request)
		if !valid {
			response := []byte("Malformed Request.\n")
			conn.Write(response)
			break

		}
		m := map[string]any{"method": "isPrime", "prime": isPrime(int64(num))}
		response, err := json.Marshal(m)
		if err != nil {
			log.Panicln("Could not marshal json: ", err.Error())
		}
		log.Printf("Responding: %s\n", response)
		conn.Write(append(response, byte('\n')))
		//		conn.Write([]byte("Hello"))
	}
	log.Println("closed connection from ", conn.RemoteAddr())
}

// validateRequest consumes a []byte representing a JSON request and returns true if it is valid according to the primetime protocal, and false if it is malformed
func validateRequest(request []byte) (bool, float64) {
	if !json.Valid(request) {
		log.Println("Invalid. Not valid json.")
		return false, 0
	}
	var requestMap map[string]any
	json.Unmarshal(request, &requestMap)
	log.Printf("Got data: %v\n", requestMap)

	method, ok := requestMap["method"]
	if !ok || method != "isPrime" {
		log.Println("Invalid. Doesn't have method isPrime.")
		return false, 0
	}

	number, ok := requestMap["number"]
	if !ok {
		log.Println("Invalid. Doesn't have number field.")
		return false, 0
	}

	// Remember that all numbers unmarshalled from json are float64. See: https://json-schema.org/understanding-json-schema/reference/numeric.html
	value, ok := number.(float64)
	if !ok {
		log.Println("Invalid. number Field does not contain a number.")
		log.Printf("Got the value %v instead of type %T", value, value)
		return false, 0
	}

	return true, value
}

func isPrime(n int64) bool {
	if n <= 3 {
		return n > 1
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	stop := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= stop; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}
