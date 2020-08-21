package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndhanushkodi/porxy/config"
)

func main() {
	rawconfig, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	c := config.LoadConfig(rawconfig)

	errs := configureProxy(c)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errs:
			fmt.Printf("Error (listen/handle): %+v\n", err)
		case sig := <-sigs:
			fmt.Println(sig)
			fmt.Println("Exiting porxy")
			os.Exit(1)
		}
	}

}

func configureProxy(c config.Config) chan error {
	errs := make(chan error)
	for _, listener := range c.Listeners {
		go createListener(listener, c.GetBackend(listener.Backend), errs)
	}
	return errs
}

// createListener starts listening on the configured address and port, and
// accepts incoming connections. Each accepted connection is then handled.
// TODO: consider the listener lifecycle
// TODO: consider sending clients to a different backend if one goes down
func createListener(listener config.Listener, backend config.Backend, errs chan error) {
	addressport := fmt.Sprintf("%s:%s", listener.Address, listener.Port)
	l, err := net.Listen("tcp", addressport)
	if err != nil {
		errs <- err
		return
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			errs <- err
			return
		}
		go handleConnection(conn, backend, errs)
	}
}

// handleConnection handles one connection from a client to server
func handleConnection(clientconn net.Conn, backend config.Backend, errs chan error) {
	addressport := fmt.Sprintf("%s:%s", backend.Host, backend.Port)
	serverconn, err := net.Dial("tcp", addressport)
	if err != nil {
		errs <- err
		return
	}

	proxyErrs := make(chan error, 1)
	go proxyConnection(serverconn, clientconn, proxyErrs)
	go proxyConnection(clientconn, serverconn, proxyErrs)

	// If either side of the connection has an error, close both sides
	// and log the error. Clients would then be able to retry connections
	err = <-proxyErrs
	fmt.Printf("Error in proxying connection, closing both sides of connection: %+v\n", err)
	serverconn.Close()
	clientconn.Close()
}

// proxyConnection handles reading and writing in one direction,
// either reading from the client and writing to the server
// or reading from the server and writing to the client.
func proxyConnection(connA, connB net.Conn, errs chan error) {
	for {
		aBytes := make([]byte, 1024)
		_, err := connA.Read(aBytes)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s -> %s closed connection\n", connA.LocalAddr().String(), connA.RemoteAddr().String())
			}
			errs <- err
			return
		}
		_, err = connB.Write(aBytes)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s -> %s closed connection\n", connB.LocalAddr().String(), connB.RemoteAddr().String())
			}
			errs <- err
			return
		}
	}
}
