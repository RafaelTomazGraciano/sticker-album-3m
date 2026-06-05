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

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrade:", err)
		return
	}

	onNewPeerConnected(conn)
	listenPeer(conn, "")
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

func onNewPeerConnected(conn *websocket.Conn) {
    sendNeighborList(conn)
    helloToNeighbors(conn)
}

func connectToPeer(addr string) {
    conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
    if err != nil {
        fmt.Printf("Erro ao conectar em %s: %v\n", addr, err)
        return
    }
    fmt.Printf("Conectado em %s\n", addr)

    node.mu.Lock()
    node.Neighbors[addr] = &PeerConn{Addr: addr, Conn: conn}
    node.mu.Unlock()

    listenPeer(conn, addr)
}
