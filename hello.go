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
        Type: "HELLO",
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

    addr := connAddr[conn]
	
    // salva o peer que enviou o HELLO como vizinho com ID correto
    node.Neighbors[msg.SenderPeerID] = &PeerConn{
        PeerID: msg.SenderPeerID,
        Conn:   conn,
        Addr:   addr,
    }

    // salva os endereços recebidos como backup
    for _, addr := range msg.Peers {
        node.KnownPeers = append(node.KnownPeers, addr)
    }
    node.mu.Unlock()

    fmt.Printf("Peer %s registrado em %s. Backup recebido: %v\n", msg.SenderPeerID, addr, msg.Peers)
}

