porxy
---
Proxies traffic

### Intent
I'm using porxy as a learning project to better understand networking. The
changelog section describes each iteration of porxy, and the featureset it
supports at that stage.

### Changelog
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
---

