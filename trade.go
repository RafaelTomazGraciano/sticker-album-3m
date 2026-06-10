package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func startTradeOffer() {
	msgTrade := Message{
        Type: "TRADE_OFFER",
        MessageID: uuid.NewString(),
		OriginPeerID: node.ID,
		SenderPeerID: node.ID,
		OfferSticker: offerSticker,
		WantSticker: wantSticker,
    }

	// TODO: if peerToTrade é seu vizinho envia direto, se não faz broadcast 
}

func handleTradeOffer(conn *websocket.Conn, msg Message){

}

func handleTradeAccept(conn *websocket.Conn, msg Message) {

}

func handleTradeReject(conn *websocket.Conn, msg Message) {

}

func handleTransferConfirm(conn *websocket.Conn, msg Message) {

}
