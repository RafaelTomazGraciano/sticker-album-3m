package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
	CheckOrigin: func(r * http.Request) bool {
		return true
	},
}

// map auxiliar: conn -> addr, para o handleHello conseguir preencher o Addr
var connAddr = make(map[*websocket.Conn]string)

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Error upgrade:", err)
        return
    }

    // pega o IP de quem conectou (formato "IP:porta")
    addr := r.RemoteAddr
    connAddr[conn] = addr

    onNewPeerConnected(conn)
    listenPeer(conn, addr)
}

func onNewPeerConnected(conn *websocket.Conn) {
    sendNeighborList(conn)
    helloToNeighbors(conn)
}

func listenPeer(conn *websocket.Conn, addr string) {
    defer conn.Close()

    for {
        _, raw, err := conn.ReadMessage()
        if err != nil {
            fmt.Printf("Peer %s desconectou\n", addr)
            handleDisconnect(addr)
            return
        }
        handleMessage(conn, raw)
    }
}

func handleDisconnect(addr string) {
    node.mu.Lock()
    // remove o peer que caiu de Neighbors
    for id, pc := range node.Neighbors {
        if pc.Addr == addr {
            delete(node.Neighbors, id)
            break
        }
    }
    semVizinhos := len(node.Neighbors) == 0
    node.mu.Unlock()

    if semVizinhos {
        fmt.Println("Sem vizinhos — tentando reconectar via Peers Conhecidos...")
        for _, known := range node.KnownPeers {
            if known != addr {
                go connectToPeer(known)
                return
            }
        }
        fmt.Println("Nenhum peer conhecido disponível")
    }
}

func connectToPeer(addr string) {
    conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
    if err != nil {
        fmt.Printf("Erro ao conectar em %s: %v\n", addr, err)
        return
    }
    fmt.Printf("Conectado em %s\n", addr)

    connAddr[conn] = addr // guarda antes do HELLO chegar

    listenPeer(conn, addr)
}
