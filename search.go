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
	_, jaEVizinho := node.Neighbors[msg.OriginPeerID]
	node.mu.RUnlock()

	if qty > 0 {
        if jaEVizinho {
            sendSearchHit(conn, msg)
        } else {
            // origin nao e vizinho. Conecta primeiro, depois envia o hit pela nova conexao
            addr := fmt.Sprintf("ws://%s:8080/ws", msg.OriginPeerIP)
            originalMsg := msg
            go connectToPeerAndDo(addr, func(newConn *websocket.Conn) {
                sendSearchHit(newConn, originalMsg)
            })
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
    // sou o destino?
    if msg.ReceiverPeerID == node.ID {
        // armazena quem tem a figurinha para usar no trade depois
        node.mu.Lock()
        node.SearchResults[msg.StickerID] = &PeerConn{
            PeerID: msg.SenderPeerID,
            Conn:   conn,
        }
        peerToTrade = msg.SenderPeerID
        node.mu.Unlock()

        fmt.Printf("Figurinha %s encontrada em %s!\n", msg.StickerID, msg.SenderPeerID)
        fmt.Printf("Use 'offer <FIG-XX>' para propor uma troca.\n")
        return
    }

    // nao sou o destino, roteia para o vizinho correto
    node.mu.RLock()
    dest, ok := node.Neighbors[msg.ReceiverPeerID]
    node.mu.RUnlock()

    if !ok {
        fmt.Printf("SEARCH_HIT: destino %s não é vizinho, não consigo rotear\n", msg.ReceiverPeerID)
        return
    }

    msg.SenderPeerID = node.ID
    data, err := json.Marshal(msg)
    if err != nil {
        fmt.Println("Erro ao serializar SEARCH_HIT no roteamento:", err)
        return
    }
    dest.Conn.WriteMessage(websocket.TextMessage, data)
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
