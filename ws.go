package main

import (
	"bufio"
	"github.com/namsral/flag"
	"io/ioutil"
	"log"
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

// Flag options
type FlagOptions struct {
	Http    string
	Tcp     string
	WsPath  string
	MsgPath string
	KeyPath string
}

func main() {
	o := ParseFlagOptions()
	// 0 - setup TCP connection
	// listen on all interfaces
	ln, _ := net.Listen("tcp", o.Tcp)
	go handleTCPMsg(ln)
	// 1
	fs := http.FileServer(http.Dir("./static"))
	// 2
	router := mux.NewRouter()
	router.Handle("/", fs) // handle tcpMsg

	// router.Handle("/", fs)
	router.HandleFunc(o.KeyPath+"/{key}", kvSave).Methods("POST")
	router.HandleFunc(o.KeyPath+"/{key}", kvGet).Methods("GET")
	router.HandleFunc(o.MsgPath, msgHandler).Methods("POST")
	router.HandleFunc(o.WsPath, wsHandler)
	go echo()
	log.Printf("Websocket server on port: " + o.Http)
	log.Fatal(http.ListenAndServe(o.Http, router))
}

func writer(coord string) {
	broadcast <- coord
}

func msgHandler(w http.ResponseWriter, r *http.Request) {
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

func handleTCPMsg(ln net.Listener) {
	defer ln.Close()
	log.Printf("listen for messages on port 15000")
	for {
		rw, e := ln.Accept()
		if e != nil {
			log.Fatal(e)
		}
		go handleTCPConnection(rw)
	}
}

func handleTCPConnection(c net.Conn) {
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

func ParseFlagOptions() *FlagOptions {
	o := &FlagOptions{
		Http:    ":18844",
		Tcp:     ":15000",
		WsPath:  "/ws",
		MsgPath: "/msg",
		KeyPath: "/key",
	}
	flag.StringVar(&o.Http, "http", o.Http, "HTTP address to listen to")
	flag.StringVar(&o.Tcp, "tcp", o.Tcp, "TCP address to listen to")
	flag.StringVar(&o.WsPath, "wspath", o.WsPath, "HTTP path for web socket client to connect to")
	flag.StringVar(&o.MsgPath, "msgpath", o.MsgPath, "HTTP path for sending message to")
	flag.StringVar(&o.KeyPath, "keypath", o.KeyPath, "HTTP path for get and set consul data")
	flag.Usage = func() {
		log.Printf("Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	return o
}
