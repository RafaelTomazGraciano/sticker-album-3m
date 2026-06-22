package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type PeerConn struct {
	PeerID string
	Addr   string
	Conn   *websocket.Conn
}

type Peer struct {
	ID            string
	StickerID     string
	Inventory     map[string]int
	Neighbors     map[string]*PeerConn
	KnownPeers    []string
	SeenQueries   map[string]time.Time
	SearchResults map[string]*PeerConn
	QueryRoutes   map[string]*websocket.Conn
	mu            sync.RWMutex
}

type Message struct {
	Type           string   `json:"type"`
	MessageID      string   `json:"message_id"`
	OriginPeerID   string   `json:"origin_peer_id"`
	OriginPeerIP    string   `json:"origin_peer_ip,omitempty"`
    ReceiverPeerIP  string   `json:"receiver_peer_ip,omitempty"`
	SenderPeerID   string   `json:"sender_peer_id"`
	ReceiverPeerID string   `json:"receiver_peer_id,omitempty"`
	QueryID        string   `json:"query_id,omitempty"`
	TTL            int      `json:"ttl,omitempty"`
	StickerID      string   `json:"sticker_id,omitempty"`
	OfferSticker   string   `json:"offer_sticker_id,omitempty"`
	WantSticker    string   `json:"want_sticker_id,omitempty"`
	Peers          []string `json:"peers,omitempty"`
}

type Album struct {
	Stickers map[string]int `json:"stickers"`
}
