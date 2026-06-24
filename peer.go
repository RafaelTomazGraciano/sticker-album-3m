package main

import (
    "fmt"
    "net"
    "net/http"
    "sync"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
	CheckOrigin: func(r * http.Request) bool {
		return true
	},
}

// map auxiliar: conn -> addr, para o handleHello conseguir preencher o Addr
var connAddrMu sync.RWMutex
var connAddr = make(map[*websocket.Conn]string)
var connWriteMu sync.Map

func getConnMutex(conn *websocket.Conn) *sync.Mutex {
    m, _ := connWriteMu.LoadOrStore(conn, &sync.Mutex{})
    return m.(*sync.Mutex)
}

func safeWriteMessage(conn *websocket.Conn, messageType int, data []byte) error {
    mu := getConnMutex(conn)
    mu.Lock()
    defer mu.Unlock()
    return conn.WriteMessage(messageType, data)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        printError("Error upgrade: %v", err)
        return
    }

    // pega o IP de quem conectou (formato "IP:porta")
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    addr := fmt.Sprintf(host)
    connAddrMu.Lock()
    connAddr[conn] = addr
    connAddrMu.Unlock()

    onNewPeerConnected(conn)
    listenPeer(conn, addr)
}

func onNewPeerConnected(conn *websocket.Conn) {
    sendHello(conn)
    helloToNeighbors(conn)
}

func listenPeer(conn *websocket.Conn, addr string) {
    defer conn.Close()

    for {
        _, raw, err := conn.ReadMessage()
        if err != nil {
            printSystem("Peer %s desconectou", addr)
            handleDisconnect(addr)
            return
        }
        handleMessage(conn, raw)
    }
}

func handleDisconnect(addr string) {
    node.mu.Lock()
    // remove o peer que caiu de Neighbors
    for id, pc := range node.Neighbors {
        if pc.Addr == addr {
            delete(node.Neighbors, id)
            break
        }
    }
    semVizinhos := len(node.Neighbors) == 0
    node.mu.Unlock()

    if semVizinhos {
        printWarning("Sem vizinhos — tentando reconectar via Peers Conhecidos...")
        for _, known := range node.KnownPeers {
            if known != addr {
                go connectToPeer(known)
                return
            }
        }
        printWarning("Nenhum peer conhecido disponível")
    }
}

func connectToPeer(addr string) {
    connectToPeerAndDo(addr, nil)
}

func connectToPeerAndDo(addr string, onConnected func(*websocket.Conn)) {
    conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
    if err != nil {
        printError("Erro ao conectar em %s: %v", addr, err)
        return
    }
    printSystem("Conectado em %s", addr)

    connAddrMu.Lock()
    connAddr[conn] = addr
    connAddrMu.Unlock() 

    if onConnected != nil {
        onConnected(conn)
    }

    listenPeer(conn, addr)
}
