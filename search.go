package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func searchWithRetry() {
    // limpa o channel de qualquer sinal anterior
    select {
    case <-searchDone:
    default:
    }

    ttl := 7
    startSearch(ttl)

    delays := []time.Duration{5 * time.Second, 10 * time.Second, 15 * time.Second}
    ttlIncrements := []int{9, 11}

    for i, delay := range delays {
        select {
        case <-searchDone:
            return
        case <-time.After(delay):
            if i < len(ttlIncrements) {
                ttl = ttlIncrements[i]
                printWarning("Sem resposta, reenviando busca...")
                startSearch(ttl)
            } else {
                printWarning("Nenhuma resposta recebida")
				printMenu()
            }
        }
    }
}

func startSearch(ttl int) {
	msg := Message{
		Type:         "SEARCH",
		MessageID:    uuid.NewString(),
		OriginPeerID: node.ID,
		OriginPeerIP: getLocalIP(),
		SenderPeerID: node.ID,
		QueryID:      uuid.NewString(),
		TTL:          ttl,
		StickerID:    wantSticker,
	}

	// registra o query_id para não processar o próprio flood
	node.mu.Lock()
	node.SeenQueries[msg.QueryID] = time.Now()
	node.mu.Unlock()

	broadcast(msg, nil)
	printInfo("Busca iniciada por %s (query: %s)", wantSticker, msg.QueryID)
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
	qty := inventory.Stickers[msg.StickerID]
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
    } else{
		sendSearchMiss(conn, msg)
	}

	// ainda tem TTL para repassar?
	if msg.TTL > 1 {
		msg.TTL--
		msg.SenderPeerID = node.ID
		broadcast(msg, conn) // repassa para todos exceto quem enviou
	}
}

func sendSearchMiss(conn *websocket.Conn, original Message) {
    miss := Message{
        Type:           "SEARCH_MISS",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: original.SenderPeerID,
        QueryID:        original.QueryID,
        StickerID:      original.StickerID,
    }

    data, err := json.Marshal(miss)
    if err != nil {
        printError("Erro ao serializar SEARCH_MISS: %v", err)
        return
    }
    conn.WriteMessage(websocket.TextMessage, data)
}

func handleSearchMiss(conn *websocket.Conn, msg Message) {
    // só loga se formos o origin da busca, senão é só ruído
    if msg.ReceiverPeerID == node.ID {
        printInfo("Nó %s não possui %s.", msg.OriginPeerID, msg.StickerID)
    }
}

func broadcast(msg Message, except *websocket.Conn) {
	data, err := json.Marshal(msg)
	if err != nil {
		printError("Erro ao serializar broadcast: %v", err)
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

		select {
			case searchDone <- struct{}{}:
			default:
				return // resposta duplicada ou tardia ignora
			}

        // armazena quem tem a figurinha para usar no trade depois
        node.mu.Lock()
        node.SearchResults[msg.StickerID] = &PeerConn{
            PeerID: msg.SenderPeerID,
            Conn: conn,
        }
        peerToTrade = msg.SenderPeerID
        node.mu.Unlock()

        printSuccess("Figurinha %s encontrada em %s!", msg.StickerID, msg.SenderPeerID)
        printSuccess("Use 'offer <FIG-XX>' para propor uma troca.")
        return
    }

    // nao sou o destino, roteia para o vizinho correto
    node.mu.RLock()
    dest, ok := node.Neighbors[msg.ReceiverPeerID]
    node.mu.RUnlock()

    if !ok {
        printWarning("SEARCH_HIT: destino %s não é vizinho, não consigo rotear", msg.ReceiverPeerID)
        return
    }

    msg.SenderPeerID = node.ID
    data, err := json.Marshal(msg)
    if err != nil {
        printError("Erro ao serializar SEARCH_HIT no roteamento: %v", err)
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
		printError("Erro ao serializar SEARCH_HIT: %v", err)
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
}
