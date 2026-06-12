package main

import (
    "fmt"
    "net"
)

var node *Peer
var localIP string

var wantSticker string
var offerSticker string

var peerToTrade string

var inventory Album
const inventoryFile = "inventory.json"

// canal para o inputLoop comunicar a decisao de accept/reject ao handleTradeOffer
var tradeDecision = make(chan string, 1)

var searchDone = make(chan struct{}, 1)

const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
    ColorBlue   = "\033[38;2;30;144;255m"
    ColorCyan   = "\033[36m"
)

func getLocalIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "127.0.0.1"
    }
    defer conn.Close()
    return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func printMenu() {
    printInfo("Comandos disponíveis:")
    printInfo("  search <FIG-XX>   -> buscar uma figurinha")
    printInfo("  offer <FIG-XX>    -> oferece uma figurinha sua para troca")
    printInfo("  accept            -> aceita a troca entre figurinhas")
    printInfo("  reject            -> rejeita a troca entre figurinhas")
    printInfo("  list              -> ver seu inventário")
    fmt.Print("> ")
}

func printSuccess(format string, args ...any) {
    fmt.Printf(ColorGreen+format+ColorReset+"\n", args...)
}

func printError(format string, args ...any) {
    fmt.Printf(ColorRed+format+ColorReset+"\n", args...)
}

func printWarning(format string, args ...any) {
    fmt.Printf(ColorYellow+format+ColorReset+"\n", args...)
}

func printInfo(format string, args ...any) {
    fmt.Printf(ColorBlue+format+ColorReset+"\n", args...)
}

func printSystem(format string, args ...any) {
    fmt.Printf(ColorCyan+format+ColorReset+"\n", args...)
}