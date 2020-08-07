porxy
---
Proxies traffic

### Intent
I'm using porxy as a learning project to better understand networking and
connection handling. The
changelog section describes each iteration of porxy, and the featureset it
supports at that stage.

### Usage
1. Build the porxy binary
   ```
   git clone https://github.com/ndhanushkodi/porxy
   cd porxy
   go build .
   ```
1. Create/Edit a configuration file called `config.yaml` in the same directory
   as the binary. See [Changelog #2](#2-configurable-connections) for the config
   spec.

1. Run clients to connect to the `listeners` in different terminal windows.

   Example TCP clients:
   ```
   nc localhost 8000
   
   nc localhost 7000
   ```

1. Run servers that the `backends` will connect to in different terminal windows.

   Example TCP servers:
   ```
   socat TCP-LISTEN:1234,crlf,reuseaddr,fork -
   
   socat TCP-LISTEN:5555,crlf,reuseaddr,fork -
   ```

1. Proxy traffic between them
   ```
   ./porxy
   ```

1. Write data on the client and server sides and see the traffic get proxied

### Changelog
---
#### 2. Configurable connections
Porxy can now be configured with a set of `listeners` and `backends`, for TCP
connections.
   Example supported config:
   ```yaml
    ---
    listeners:
      - name: moo
        backend: foo
        address: 0.0.0.0
        port: 8000
      - name: roo
        backend: bar
        address: 0.0.0.0
        port: 7000

    backends:
      - name: foo
        host: localhost
        port: 1234
      - name: bar
        host: localhost
        port: 5555
   ```
In this example, porxy will proxy traffic from port 8000 to a server on
`localhost:1234` and from port 7000 to a server on `localhost:5555`.

The goal of this changeset was to just get configurable connections working, but
it's becoming clear there needs to be a way to handle/log errors for individual
connections, so that is likely to be the next changeset.

---
#### 1. Single connection TCP Proxy
Porxy listens for a TCP client connection on a hardcoded port. When porxy receives a
connection, it then connects to a hardcoded port that has a TCP server listening,
and forwards any data on the client side to the server and any data on the
server side to the client.

Example TCP server:
`socat TCP-LISTEN:1234,crlf,reuseaddr,fork -`

Example TCP client:
`nc localhost 8000`

Connection handling
* When a client disconnects, porxy's connection to both sides should be cleaned up
  without affecting the server
* When a server goes offline, the client should behave the same way it would if
  porxy didn't exist. The client side exits when it sends more data over the
  broken connection and receives no ACK from the server.

##### Interestings
The following github issues describe that reading from a `net.Conn` is blocking:
* [Waiting on readability for TCPConn](https://github.com/golang/go/issues/15735#issuecomment-266574151)
* [Non-blocking read on net.Conn](https://github.com/golang/go/issues/36973)

These resources could be an interesting follow up exercise, to explore whether porxy performance improves using asynchronous syscalls to read/write from the socket, rather than using the net.Conn library.
* [Asynchronous system calls](https://thenewstack.io/how-io_uring-and-ebpf-will-revolutionize-programming-in-linux/)
* [go mysql driver raw connection handling](https://github.com/go-sql-driver/mysql/blob/master/conncheck.go)
* [go async io library](https://github.com/xtaci/gaio)

---

