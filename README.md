# Golang Websocket

This app is a simple websocket server that have the following function:

*  Run a HTTP server in port `18844`
*  Listen for websocket connection in `/ws` endpoint. Browser can connect to it via `ws://IP:18844/ws` endpoint.
*  Other app can send HTTP POST to `/msg` endpoint and the websocket server will broadcast the message to its client.

## Build and Run


```
$ go build ws.go
$ ./ws
```

With bundle


```
$ go get -u github.com/go-bindata/go-bindata/...
$ ~/go/bin/go-bindata -fs -prefix "static/" static/
$ go build ws.go bindata.go
```
