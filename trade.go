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
    qty := inventory.Stickers[offerSticker]
    result, temResultado := node.SearchResults[wantSticker]
    node.mu.RUnlock()

    if qty <= 0 {
        printWarning("Você não possui '%s' para oferecer", offerSticker)
        fmt.Print("> ")
        return
    }

    if !temResultado {
        printWarning("Nenhum resultado de busca para %s. Faça um 'search' primeiro.", wantSticker)
        fmt.Print("> ")
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
        printError("Erro ao serializar TRADE_OFFER: %v", err)
        return
    }

    if result.Conn != nil {
        // já tem conexão direta (vizinho direto)
        result.Conn.WriteMessage(websocket.TextMessage, data)
        printInfo("Oferta enviada diretamente para %s", peerToTrade)
    } else if result.Addr != "" {
        // não é vizinho direto: conecta pelo IP salvo no SEARCH_HIT
        msgCopy := msgTrade
        go connectToPeerAndDo(result.Addr, func(newConn *websocket.Conn) {
            newConn.WriteMessage(websocket.TextMessage, data)
            printInfo("Oferta enviada via nova conexão para %s", msgCopy.ReceiverPeerID)
        })
    } else {
        printWarning("Sem rota para %s. Tente buscar novamente.", peerToTrade)
    }
}

func handleTradeOffer(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        return
    }

    node.mu.RLock()
    qty := inventory.Stickers[msg.WantSticker]
    node.mu.RUnlock()

    if qty <= 0 {
        printWarning("Proposta de %s recebida, mas você não possui '%s'. Rejeitando automaticamente", msg.OriginPeerID, msg.WantSticker)
        sendTradeReject(conn, msg)
        return
    }

    printWarning("\n--- PROPOSTA DE TROCA ---")
    printWarning(" De: %s", msg.OriginPeerID)
    printWarning(" Oferece: %s", msg.OfferSticker)
    printWarning(" Quer: %s", msg.WantSticker)
    printWarning(" Digite 'accept' para aceitar ou 'reject' para recusar")
    fmt.Print("> ")

    select {
    case decision := <-tradeDecision:
        if decision == "accept" {
            sendTradeAccept(conn, msg)
        } else {
            sendTradeReject(conn, msg)
        }
    case <-time.After(30 * time.Second):
        printWarning("Tempo esgotado. Rejeitando proposta automaticamente")
        sendTradeReject(conn, msg)
		printMenu()
    }
}

func sendTradeAccept(conn *websocket.Conn, original Message) {
    msg := Message{
        Type:           "TRADE_ACCEPT",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: original.OriginPeerID,
        OfferSticker:   original.WantSticker,
        WantSticker:    original.OfferSticker,
    }
    
    data, _ := json.Marshal(msg)
    conn.WriteMessage(websocket.TextMessage, data)

    updateInventory(original.OfferSticker, original.WantSticker)
    printSuccess("Troca aceita com %s. +%s / -%s", original.OriginPeerID, original.OfferSticker, original.WantSticker)
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
    printError("Troca recusada")
}

func handleTradeAccept(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        return
    }

	printSuccess("%s aceitou a troca! Confirmando transferência...", msg.OriginPeerID)

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
        printError("Erro ao serializar TRANSFER_CONFIRM: %v", err)
        return
    }
    conn.WriteMessage(websocket.TextMessage, data)

    // atualiza o inventario
    updateInventory(msg.OfferSticker, msg.WantSticker)
    printSuccess("Inventário atualizado: +%s / -%s", msg.OfferSticker, msg.WantSticker)
}

func handleTradeReject(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        return
    }

	printError("%s recusou a troca", msg.OriginPeerID)
}

func handleTransferConfirm(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        return
    }
    printSuccess("Transferência confirmada por %s!", msg.OriginPeerID)
}
