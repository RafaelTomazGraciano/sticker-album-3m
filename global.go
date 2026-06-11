package main

import "net"

var node *Peer

var wantSticker string
var offerSticker string

var peerToTrade string

var inventory Album
const inventoryFile = "inventory.json"

// canal para o inputLoop comunicar a decisao de accept/reject ao handleTradeOffer
var tradeDecision = make(chan string, 1)


func getLocalIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "127.0.0.1"
    }
    defer conn.Close()
    return conn.LocalAddr().(*net.UDPAddr).IP.String()
}