package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func startSearch() {
	msg := Message{
		Type:         "SEARCH",
		MessageID:    uuid.NewString(),
		OriginPeerID: node.ID,
		OriginPeerIP: getLocalIP(),
		SenderPeerID: node.ID,
		QueryID:      uuid.NewString(),
		TTL:          7,
		StickerID:    wantSticker,
	}

	// registra o query_id para não processar o próprio flood
	node.mu.Lock()
	node.SeenQueries[msg.QueryID] = time.Now()
	node.mu.Unlock()

	broadcast(msg, nil)
	fmt.Printf("Busca iniciada por %s (query: %s)\n", wantSticker, msg.QueryID)
}

func handleSearch(conn *websocket.Conn, msg Message) {
	// ja processou esse query_id?
	node.mu.Lock()
	if _, seen := node.SeenQueries[msg.QueryID]; seen {
		node.mu.Unlock()
		return
	}
	node.SeenQueries[msg.QueryID] = time.Now()
	node.mu.Unlock()

	// tenho a figurinha?
	node.mu.RLock()
	qty := node.Inventory[msg.StickerID]
	node.mu.RUnlock()

	// tenta se conectar para virar vizinho
	if qty > 0 {
		if _, ok := node.Neighbors[msg.OriginPeerID]; !ok {
			addr := fmt.Sprintf("ws://%s:8080/ws", msg.OriginPeerIP)
			go connectToPeer(addr)
		} else {
			sendSearchHit(conn, msg)
		}
	}

	// ainda tem TTL para repassar?
	if msg.TTL > 1 {
		msg.TTL--
		msg.SenderPeerID = node.ID
		broadcast(msg, conn) // repassa para todos exceto quem enviou
	}
}

func broadcast(msg Message, except *websocket.Conn) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Erro ao serializar broadcast:", err)
		return
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	for _, pc := range node.Neighbors {
		if pc.Conn != except {
			pc.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func handleSearchHit(conn *websocket.Conn, msg Message) {
	peerToTrade = msg.SenderPeerID
	fmt.Printf("Figurinha %s encontrada em %s!\n", msg.StickerID, msg.SenderPeerID)

    //TODO:
    //1. Se eu não sou o alvo do search hit devo passar para os próximos
    //2. Se eu sou o alvo, eu devo ser capaz de armazenar o ID de quem tem, para isso devo saber se o ip que mando
}

func sendSearchHit(conn *websocket.Conn, original Message) {
	hit := Message{
		Type:           "SEARCH_HIT",
		MessageID:      uuid.NewString(),
		OriginPeerID:   node.ID,
		SenderPeerID:   node.ID,
		ReceiverPeerID: original.OriginPeerID,
		QueryID:        original.QueryID,
		StickerID:      original.StickerID,
	}

	data, err := json.Marshal(hit)
	if err != nil {
		fmt.Println("Erro ao serializar SEARCH_HIT:", err)
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
}
