package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/ndhanushkodi/porxy/config"
)

func main() {
	rawconfig, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	c := config.LoadConfig(rawconfig)

	for _, listener := range c.Listeners {
		go createListener(listener, c.GetBackend(listener.Backend))
	}
	for {
		//TODO if all listeners exit, porxy could close
		//TODO if porxy gets a signal, it should log it and close
		//TODO bubble up errors in an error channel and log them and exit
		time.Sleep(1 * time.Second)
	}

}

func createListener(listener config.Listener, backend config.Backend) {
	addressport := fmt.Sprintf("%s:%s", listener.Address, listener.Port)
	l, err := net.Listen("tcp", addressport)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn, backend)
	}
}

// handleConnection handles one connection from a client to server
func handleConnection(clientconn net.Conn, backend config.Backend) {
	addressport := fmt.Sprintf("%s:%s", backend.Host, backend.Port)
	serverconn, err := net.Dial("tcp", addressport)
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
