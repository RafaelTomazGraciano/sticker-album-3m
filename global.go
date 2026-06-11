package main

var node *Peer

var wantSticker string
var offerSticker string

var peerToTrade string

var inventory Album
const inventoryFile = "inventory.json"


func getLocalIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "127.0.0.1"
    }
    defer conn.Close()
    return conn.LocalAddr().(*net.UDPAddr).IP.String()
}