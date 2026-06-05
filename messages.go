package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

func handleMessage(conn *websocket.Conn, raw []byte) {
    var msg Message
    if err := json.Unmarshal(raw, &msg); err != nil {
        fmt.Println("Erro ao parsear mensagem:", err)
        return
    }

    switch msg.Type {
    case "HELLO":
        handleHello(conn, msg)
    case "SEARCH":
        handleSearch(conn, msg)
    case "SEARCH_HIT":
        handleSearchHit(conn, msg)
    case "TRADE_OFFER":
        handleTradeOffer(conn, msg)
    case "TRADE_ACCEPT":
        handleTradeAccept(conn, msg)
    case "TRADE_REJECT":
        handleTradeReject(conn, msg)
    case "TRANSFER_CONFIRM":
        handleTransferConfirm(conn, msg)
    default:
        fmt.Printf("Tipo de mensagem desconhecido: %s\n", msg.Type)
    }
}