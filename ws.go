package main

import (
	"bufio"
	"log"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var kvStorage = make(map[string]string)

func main() {
	// 0 - setup TCP connection
	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":15000")
	go handleTcpMsg(ln)
	// 1
	fs := http.FileServer(http.Dir("./static"))
	// 2
	router := mux.NewRouter()
	router.Handle("/", fs)// handle tcpMsg

	// router.Handle("/", fs)
	router.HandleFunc("/kv/{key}", kvSave).Methods("POST")
	router.HandleFunc("/kv/{key}", kvGet).Methods("GET")
	router.HandleFunc("/msg", longLatHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	router.HandleFunc("/ws-hanoi", wsHandler)
	router.HandleFunc("/ws-hcm", wsHandler)
	go echo()
	log.Printf("Websocket server on port: :18844")
	log.Fatal(http.ListenAndServe(":18844", router))
}

func writer(coord string) {
	broadcast <- coord
}

func longLatHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	go writer(string(body))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("New client connected")
	// register client
	clients[ws] = true
}

func kvSave(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	key := vars["key"]
	body, _ := ioutil.ReadAll(r.Body)
	kvStorage[key] = string(body)
	w.Write(body)
}

func kvGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	key := vars["key"]
	w.Write([]byte(kvStorage[key]))
}

func handleTcpMsg(ln net.Listener) {
	defer ln.Close()
	log.Printf("listen for messages on port 15000")
	for {
			rw, e := ln.Accept()
			if e != nil {
					log.Fatal(e)
			}
			go handleTcpConnection(rw)
	}
}

func handleTcpConnection(c net.Conn) {
	log.Printf("New connection established: %s", c.RemoteAddr().String())
	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			// handle error
			return
		}
		go writer(message)
	}
	c.Close()
}
// 3
func echo() {
	for {
		val := <-broadcast
		// latlong := fmt.Sprintf("%f %f %s", val.Lat, val.Long)
		// send to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(val))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
