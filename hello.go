package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

func sendNeighborList(conn *websocket.Conn) {
    node.mu.RLock()
    neighbors := make([]string, 0, len(node.Neighbors))
    for _, pc := range node.Neighbors {
        neighbors = append(neighbors, pc.Addr)
    }
    node.mu.RUnlock()

    msg := Message{
        Type:         "HELLO",
        SenderPeerID: node.ID,
        Peers:        neighbors,
    }

    data, err := json.Marshal(msg)
    if err != nil {
        fmt.Println("Erro ao serializar HELLO:", err)
        return
    }
    conn.WriteMessage(websocket.TextMessage, data)
}

func helloToNeighbors(newConn *websocket.Conn) {
    node.mu.RLock()
    defer node.mu.RUnlock()

    msg := Message{
        Type:         "HELLO",
        SenderPeerID: node.ID,
    }

    data, err := json.Marshal(msg)
    if err != nil {
        fmt.Println("Erro ao serializar HELLO:", err)
        return
    }

    for _, pc := range node.Neighbors {
        if pc.Conn != newConn {
            pc.Conn.WriteMessage(websocket.TextMessage, data)
        }
    }
}

func handleHello(conn *websocket.Conn, msg Message) {
    node.mu.Lock()
    for _, addr := range msg.Peers {
        node.KnownPeers = append(node.KnownPeers, addr)
    }
    node.mu.Unlock()

    fmt.Printf("Backup de peers recebido: %v\n", msg.Peers)
}

