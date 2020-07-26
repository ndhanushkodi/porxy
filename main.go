package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	// Listen on port 8000 for netcat client
	// "nc localhost 8000"
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn)
	}
}

// handleConnection handles one connection from a client to server
func handleConnection(clientconn net.Conn) {
	// Connect to socat server listening on 1234
	// "socat TCP-LISTEN:1234,crlf,reuseaddr,fork -"
	serverconn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		panic(err)
	}

	errChan := make(chan error, 1)
	go proxyConnection(serverconn, clientconn, errChan)
	go proxyConnection(clientconn, serverconn, errChan)

	// If either side of the connection has an error, close both sides
	// and log the error. Clients would then be able to retry connections
	err = <-errChan
	fmt.Println(err)
	serverconn.Close()
	clientconn.Close()
}

// proxyConnection handles reading and writing in one direction,
// either reading from the client and writing to the server
// or reading from the server and writing to the client.
func proxyConnection(connA, connB net.Conn, errChan chan error) {
	for {
		aBytes := make([]byte, 1024)
		_, err := connA.Read(aBytes)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s -> %s closed connection\n", connA.LocalAddr().String(), connA.RemoteAddr().String())
			}
			errChan <- err
			return
		}
		_, err = connB.Write(aBytes)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s -> %s closed connection\n", connB.LocalAddr().String(), connB.RemoteAddr().String())
			}
			errChan <- err
			return
		}
	}
}
