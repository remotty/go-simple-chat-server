package main

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
)

type ClientConn struct {
	websocket *websocket.Conn
	clientIP  net.Addr
}

var (
	ActiveClients        = make(map[ClientConn]string)
	ActiveClientsRWMutex sync.RWMutex
)

func init() {
	// init global variables
}

func main() {
	r := mux.NewRouter().StrictSlash(false)

	// routing websocket
	r.Path("/ws/chat/{key}").HandlerFunc(wsHandler)

	n := negroni.Classic()
	n.UseHandler(r)

	n.Run(":4000")
}

// websocket handler

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(ActiveClients)

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	key := mux.Vars(r)["key"]
	client := ws.RemoteAddr()
	sockCli := ClientConn{ws, client}
	addClient(key, sockCli)

	for {
		log.Println(len(ActiveClients), ActiveClients)
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			deleteClient(sockCli)
			log.Println("bye")
			log.Println(err)
			return
		}
		broadcastMessage(key, messageType, p)
	}
}

func addClient(key string, cc ClientConn) {
	ActiveClientsRWMutex.Lock()
	ActiveClients[cc] = key
	ActiveClientsRWMutex.Unlock()
}

func deleteClient(cc ClientConn) {
	ActiveClientsRWMutex.Lock()
	delete(ActiveClients, cc)
	ActiveClientsRWMutex.Unlock()
}

func broadcastMessage(key string, messageType int, message []byte) {
	ActiveClientsRWMutex.RLock()
	defer ActiveClientsRWMutex.RUnlock()

	for client, k := range ActiveClients {
		log.Println(client)
		log.Println(k)
		if key != k {
			continue
		}
		if err := client.websocket.WriteMessage(messageType, message); err != nil {
			continue
		}
	}
}
