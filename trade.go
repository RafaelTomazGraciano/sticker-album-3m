package main

import (
	"github.com/gorilla/websocket"
)

func handleTradeOffer(conn *websocket.Conn, msg Message)      {}
func handleTradeAccept(conn *websocket.Conn, msg Message)     {}
func handleTradeReject(conn *websocket.Conn, msg Message)     {}
func handleTransferConfirm(conn *websocket.Conn, msg Message) {}