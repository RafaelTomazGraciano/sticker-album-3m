package main

import (
	"fmt"
	"sync"
	"time"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var pendingOffer *Message
var pendingOfferMu sync.RWMutex

func setPendingOffer(msg *Message) {
    pendingOfferMu.Lock()
    pendingOffer = msg
    pendingOfferMu.Unlock()
}

func getPendingOffer() *Message {
    pendingOfferMu.RLock()
    defer pendingOfferMu.RUnlock()
    return pendingOffer
}

func startTradeOffer() {
    node.mu.RLock()
    offer := offerSticker
    want := wantSticker
    target := peerToTrade
    qty := inventory.Stickers[offer]
    result, temResultado := node.SearchResults[want]
    node.mu.RUnlock()

    if qty <= 0 {
        printWarning("Você não possui '%s' para oferecer", offer)
        fmt.Print("> ")
        return
    }

    if !temResultado {
        printWarning("Nenhum resultado de busca para %s. Faça um 'search' primeiro.", want)
        fmt.Print("> ")
        return
    }

    msgTrade := Message{
        Type:           "TRADE_OFFER",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: target,
        QueryID:        result.QueryID,
        OfferSticker:   offer,
        WantSticker:    want,
    }

    data, err := json.Marshal(msgTrade)
    if err != nil {
        printError("Erro ao serializar TRADE_OFFER: %v", err)
        return
    }

    if result.Conn != nil {
        // já tem conexão direta (vizinho direto)
        safeWriteMessage(result.Conn, websocket.TextMessage, data)
        printInfo("Oferta enviada diretamente para %s", target)
    } else if result.Addr != "" {
        // não é vizinho direto: conecta pelo IP salvo no SEARCH_HIT
        msgCopy := msgTrade
        go connectToPeerAndDo(result.Addr, func(newConn *websocket.Conn) {
            safeWriteMessage(newConn, websocket.TextMessage, data)
            printInfo("Oferta enviada via nova conexão para %s", msgCopy.ReceiverPeerID)
        })
    } else {
        printWarning("Sem rota para %s. Tente buscar novamente.", target)
    }
}

func handleTradeOffer(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        // não é pra mim: repassa na direção de quem tem a figurinha
        node.mu.RLock()
        forward, temRota := node.QueryForwardRoute[msg.QueryID]
        node.mu.RUnlock()

        if !temRota || forward == nil {
            printWarning("TRADE_OFFER para %s sem rota conhecida, descartando", msg.ReceiverPeerID)
            return
        }

        data, err := json.Marshal(msg)
        if err != nil {
            printError("Erro ao serializar TRADE_OFFER no roteamento: %v", err)
            return
        }
        safeWriteMessage(forward, websocket.TextMessage, data)
        return
    }

    tradeMu.Lock()
    if tradeInProgress {
        tradeMu.Unlock()
        printWarning("Proposta de %s recebida, mas já há uma negociação em andamento. Rejeitando automaticamente", msg.OriginPeerID)
        sendTradeReject(conn, msg)
        return
    }
    tradeInProgress = true
    tradeMu.Unlock()

    defer func() {
        tradeMu.Lock()
        tradeInProgress = false
        tradeMu.Unlock()
    }()

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

    setPendingOffer(&msg)

    select {
    case decision := <-tradeDecision:
        setPendingOffer(nil)
        if decision == "accept" {
            sendTradeAccept(conn, msg)
        } else {
            sendTradeReject(conn, msg)
        }
    case <-time.After(30 * time.Second):
        setPendingOffer(nil)
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
        QueryID:        original.QueryID,
        OfferSticker:   original.WantSticker,
        WantSticker:    original.OfferSticker,
    }

    data, _ := json.Marshal(msg)
    safeWriteMessage(conn, websocket.TextMessage, data)

    printInfo("Troca aceita, aguardando confirmação de %s...", original.OriginPeerID)
}

func sendTradeReject(conn *websocket.Conn, original Message) {
    msg := Message{
        Type:           "TRADE_REJECT",
        MessageID:      uuid.NewString(),
        OriginPeerID:   node.ID,
        SenderPeerID:   node.ID,
        ReceiverPeerID: original.OriginPeerID,
        QueryID:        original.QueryID,
        OfferSticker:   original.OfferSticker,
        WantSticker:    original.WantSticker,
    }
    data, _ := json.Marshal(msg)
    safeWriteMessage(conn, websocket.TextMessage, data)
    printError("Troca recusada")
}

func handleTradeAccept(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        // não é pra mim: repassa na direção de quem iniciou a busca
        node.mu.RLock()
        route, temRota := node.QueryRoutes[msg.QueryID]
        node.mu.RUnlock()

        if !temRota || route == nil {
            printWarning("TRADE_ACCEPT para %s sem rota conhecida, descartando", msg.ReceiverPeerID)
            return
        }

        data, err := json.Marshal(msg)
        if err != nil {
            printError("Erro ao serializar TRADE_ACCEPT no roteamento: %v", err)
            return
        }
        safeWriteMessage(route, websocket.TextMessage, data)
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
        QueryID:        msg.QueryID,
        OfferSticker:   msg.OfferSticker,
        WantSticker:    msg.WantSticker,
    }

    data, err := json.Marshal(confirm)
    if err != nil {
        printError("Erro ao serializar TRANSFER_CONFIRM: %v", err)
        return
    }
    safeWriteMessage(conn, websocket.TextMessage, data)

    // atualiza o inventario
    updateInventory(msg.OfferSticker, msg.WantSticker)
    printSuccess("Inventário atualizado: +%s / -%s", msg.OfferSticker, msg.WantSticker)
}

func handleTradeReject(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        node.mu.RLock()
        route, temRota := node.QueryRoutes[msg.QueryID]
        node.mu.RUnlock()

        if !temRota || route == nil {
            printWarning("TRADE_REJECT para %s sem rota conhecida, descartando", msg.ReceiverPeerID)
            return
        }

        data, err := json.Marshal(msg)
        if err != nil {
            printError("Erro ao serializar TRADE_REJECT no roteamento: %v", err)
            return
        }
        safeWriteMessage(route, websocket.TextMessage, data)
        return
    }

	printError("%s recusou a troca", msg.OriginPeerID)
}

func handleTransferConfirm(conn *websocket.Conn, msg Message) {
    if msg.ReceiverPeerID != "" && msg.ReceiverPeerID != node.ID {
        node.mu.RLock()
        forward, temRota := node.QueryForwardRoute[msg.QueryID]
        node.mu.RUnlock()

        if !temRota || forward == nil {
            printWarning("TRANSFER_CONFIRM para %s sem rota conhecida, descartando", msg.ReceiverPeerID)
            return
        }

        data, err := json.Marshal(msg)
        if err != nil {
            printError("Erro ao serializar TRANSFER_CONFIRM no roteamento: %v", err)
            return
        }
        safeWriteMessage(forward, websocket.TextMessage, data)
        return
    }

    updateInventory(msg.WantSticker, msg.OfferSticker)
    printSuccess("Transferência confirmada por %s! +%s / -%s", msg.OriginPeerID, msg.WantSticker, msg.OfferSticker)
}
