package main

import (
	"io/ioutil"
	"log"
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

func main() {
	// 1
	// fs := http.FileServer(http.Dir("./static"))
	// 2
	router := mux.NewRouter()
	// router.Handle("/", fs)
	router.Handle("/", http.FileServer(AssetFile()))
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
