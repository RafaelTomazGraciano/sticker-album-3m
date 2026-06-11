package main

import (
	"fmt"
	"time"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func startTradeOffer() {
    node.mu.RLock()
    qty := node.Inventory[offerSticker]
    result, temConexaoDireta := node.SearchResults[wantSticker]
    node.mu.RUnlock()

    if qty <= 0 {
        fmt.Printf("Você não possui '%s' para oferecer.\n", offerSticker)
        return
    }

    msgTrade := Message{
        Type:           "TRADE_OFFER",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: peerToTrade,
        OfferSticker:   offerSticker,
        WantSticker:    wantSticker,
    }

    data, err := json.Marshal(msgTrade)
    if err != nil {
        fmt.Println("Erro ao serializar TRADE_OFFER:", err)
        return
    }

    if temConexaoDireta {
        result.Conn.WriteMessage(websocket.TextMessage, data)
        fmt.Printf("Oferta enviada diretamente para %s.\n", peerToTrade)
    } else {
        broadcast(msgTrade, nil)
        fmt.Printf("Oferta enviada por broadcast para %s.\n", peerToTrade)
    }
}

func handleTradeOffer(conn *websocket.Conn, msg Message) {
    node.mu.RLock()
    qty := node.Inventory[msg.WantSticker]
    node.mu.RUnlock()

    if qty <= 0 {
        fmt.Printf("Proposta de %s recebida, mas você não possui '%s'. Rejeitando automaticamente.\n", msg.OriginPeerID, msg.WantSticker)
        sendTradeReject(conn, msg)
        return
    }

    fmt.Printf("\n--- PROPOSTA DE TROCA ---\n")
    fmt.Printf("  De:      %s\n", msg.OriginPeerID)
    fmt.Printf("  Oferece: %s\n", msg.OfferSticker)
    fmt.Printf("  Quer:    %s\n", msg.WantSticker)
    fmt.Printf("  Digite 'accept' para aceitar ou 'reject' para recusar.\n")
    fmt.Print("> ")

    select {
    case decision := <-tradeDecision:
        if decision == "accept" {
            sendTradeAccept(conn, msg)
        } else {
            sendTradeReject(conn, msg)
        }
    case <-time.After(30 * time.Second):
        fmt.Println("Tempo esgotado. Rejeitando proposta automaticamente.")
        sendTradeReject(conn, msg)
    }
}

func sendTradeAccept(conn *websocket.Conn, original Message) {
    msg := Message{
        Type:           "TRADE_ACCEPT",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: original.OriginPeerID,
        OfferSticker:   original.OfferSticker,
        WantSticker:    original.WantSticker,
    }
    data, _ := json.Marshal(msg)
    conn.WriteMessage(websocket.TextMessage, data)
    fmt.Printf("Troca aceita com %s.\n", original.OriginPeerID)
}

func sendTradeReject(conn *websocket.Conn, original Message) {
    msg := Message{
        Type:           "TRADE_REJECT",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: original.OriginPeerID,
        OfferSticker:   original.OfferSticker,
        WantSticker:    original.WantSticker,
    }
    data, _ := json.Marshal(msg)
    conn.WriteMessage(websocket.TextMessage, data)
    fmt.Printf("Troca recusada.\n")
}

func handleTradeAccept(conn *websocket.Conn, msg Message) {
	fmt.Printf("\n%s aceitou a troca! Confirmando transferência...\n", msg.OriginPeerID)

    // envia CONFIRM
    confirm := Message{
        Type:           "TRANSFER_CONFIRM",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: msg.OriginPeerID,
        OfferSticker:   msg.OfferSticker,
        WantSticker:    msg.WantSticker,
    }

    data, err := json.Marshal(confirm)
    if err != nil {
        fmt.Println("Erro ao serializar TRANSFER_CONFIRM:", err)
        return
    }
    conn.WriteMessage(websocket.TextMessage, data)

    // atualiza o inventario
    updateInventory(msg.WantSticker, msg.OfferSticker)
    fmt.Printf("Inventário atualizado: +%s / -%s\n", msg.WantSticker, msg.OfferSticker)
}

func handleTradeReject(conn *websocket.Conn, msg Message) {
	fmt.Printf("\n%s recusou a troca.\n", msg.OriginPeerID)
}

func handleTransferConfirm(conn *websocket.Conn, msg Message) {
	updateInventory(msg.OfferSticker, msg.WantSticker)
    fmt.Printf("Transferência confirmada! +%s / -%s\n", msg.OfferSticker, msg.WantSticker)
}
